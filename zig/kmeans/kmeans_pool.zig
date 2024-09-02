const std = @import("std");
const rand = std.crypto.random;
const print = std.debug.print;

const DIMS = 8; // dimension of the vectors we are working with
const vecOps = VectorOps(DIMS, f32);
pub fn run(K: usize, file_name: []const u8) !void {
    // var gpa = std.heap.GeneralPurposeAllocator(.{ .thread_safe = true }){};
    // defer _ = gpa.deinit();
    // var gpa_allocator = gpa.allocator();
    // var arena = std.heap.ArenaAllocator.init(gpa.allocator());
    // _ = arena;
    const file = try std.fs.cwd().openFile(file_name, .{});
    var buffered = std.io.bufferedReader(file.reader());
    var reader = buffered.reader();
    defer file.close();

    var allocator = std.heap.c_allocator;
    var thread_pool: std.Thread.Pool = undefined;
    const cpuCount = try std.Thread.getCpuCount();
    std.debug.print("cpu count: {d}\n", .{cpuCount});
    try thread_pool.init(.{ .allocator = allocator, .n_jobs = cpuCount });
    defer thread_pool.deinit();

    // precision to use to determine if the centroids moved
    const epsilon: f32 = 0.01;
    // const K: usize = arg; // number of clusters to build

    // kmeans_groups is the final data structure that holds the mapping of
    // centroids to elements.
    // var kmeans_groups = LinkedList(@Vector(DIMS, f32)).init(&allocator);
    // defer kmeans_groups.removeAll();

    // clusters is a list of lists for holding centroids -> members of the group
    var clusters = TSA(TSA(@Vector(DIMS, f32))).init(&allocator);
    defer {
        // Deinitialize each array list inside the clusters array list
        for (clusters.list.items) |*cluster| {
            cluster.deinit();
        }
        clusters.deinit();
    }

    // std in reader and wrapper to pipe in vector file
    // const stdin = std.io.getStdIn();
    // const stdinReader = file.reader();

    // buffer and wrapped stream for reading in the vectors from a file
    // var buf = std.ArrayList(u8).init(allocator);
    var buf: [1024 * 1024]u8 = undefined;
    var writer = std.io.fixedBufferStream(&buf);

    // vector data from file
    var vecData = TSA(@Vector(DIMS, f32)).init(&allocator);
    try vecData.list.ensureTotalCapacity(1000000);
    defer vecData.deinit();

    var start = std.time.milliTimestamp();
    while (true) {
        defer writer.reset();
        // read the file line by line writing into buf
        reader.streamUntilDelimiter(writer.writer(), '\n', null) catch |err| {
            switch (err) {
                error.EndOfStream => {
                    break;
                },
                else => {
                    print("{any}", .{err});
                    break;
                },
            }
        };

        // parse the json array as a fixed size f32 array
        const arr = std.json.parseFromSlice([DIMS]f32, allocator, buf[0..writer.pos], .{}) catch |err| {
            switch (err) {
                error.UnexpectedEndOfInput => {
                    break;
                },
                else => {
                    print("{any} error reading json\n", .{err});
                    break;
                },
            }
        };

        try vecData.append(arr.value);
        defer arr.deinit();
    }
    var end = std.time.milliTimestamp();
    std.debug.print("file read time: {d}ms\n", .{end - start});

    var centroids = TSA(@Vector(DIMS, f32)).init(&allocator);
    defer centroids.deinit();

    // pick random points to use as centroids
    for (0..K) |_| {
        const d = rand.intRangeAtMost(usize, 0, vecData.list.items.len - 1);
        try centroids.append(vecData.list.items[d]);
    }

    // Initialize the arraylists that will contain the vectors for each centroid
    for (0..K) |_| {
        var arr = TSA(@Vector(DIMS, f32)).init(&allocator);
        try arr.list.ensureTotalCapacity(1000000);
        try clusters.append(arr);
    }

    var wg = std.Thread.WaitGroup{};
    start = std.time.milliTimestamp();
    while (true) {
        // loop to classify each vector into a cluster
        for (vecData.list.items) |vec| {
            wg.start();
            try thread_pool.spawn(assignCentroid, .{ &wg, centroids, vec, &clusters });
        }

        thread_pool.waitAndWork(&wg);
        wg.reset();

        var moved: bool = false;
        for (centroids.list.items, clusters.list.items) |*centroid, cluster| {
            const new_centroid = vecOps.mean(cluster);
            if (vecOps.dist(centroid.*, new_centroid) > epsilon) {
                moved = true;
            }
            centroid.* = new_centroid;
        }

        // if we did not move, then we have good enough centroids
        // we are done
        if (!moved) {
            // for (centroids.list.items, clusters.list.items) |centroid, clusters_items| {
            // _ = centroid;
            // std.debug.print("Centroid len {d}\n", .{clusters_items.list.items.len});
            // }
            break;
        }

        // clean up all clusters as well since we are going to re-calculate
        // them all on the next loop
        for (clusters.list.items) |*c| {
            c.list.clearRetainingCapacity();
        }
    }
    end = std.time.milliTimestamp();
    std.debug.print("kmeans time: {d}ms\n", .{end - start});
    // we are done
    // kmeans_groups.print();
}

fn assignCentroid(wg: *std.Thread.WaitGroup, centroids: TSA(@Vector(DIMS, f32)), vec: @Vector(DIMS, f32), clusters: *TSA(TSA(@Vector(DIMS, f32)))) void {
    defer wg.finish();
    var bestCluster: usize = 0; // the index of the cluster we assign the vector to
    var minDist: f32 = std.math.inf(f32);
    for (centroids.list.items, 0..) |centroid, i| {
        const dist: f32 = vecOps.dist(vec, centroid);
        if (dist < minDist) {
            minDist = dist;
            bestCluster = i;
        }
    }

    clusters.list.items[bestCluster].append(vec) catch |err| {
        std.debug.print("error {any}\n", .{err});
    };
}

// pub fn LinkedList(comptime T: type) type {
//     return struct {
//         const This = @This();
//         const Node = struct {
//             centroid: T,
//             members: TSA(T),
//             next: ?*Node,
//         };
//
//         allocator: *std.mem.Allocator,
//         head: ?*Node,
//         len: usize,
//
//         pub fn init(allocator: *std.mem.Allocator) This {
//             return .{
//                 .allocator = allocator,
//                 .head = null,
//                 .len = 0,
//             };
//         }
//
//         pub fn append(self: *This, centroid: T, members: TSA(T)) !void {
//             const new_node: *Node = try self.allocator.create(Node);
//             new_node.* = Node{ .centroid = centroid, .members = members, .next = self.head };
//             self.head = new_node;
//             self.len += 1;
//         }
//
//         pub fn removeAll(self: *This) void {
//             var current_node: ?*Node = self.head;
//             while (current_node) |node| {
//                 current_node = node.next;
//
//                 node.members.deinit();
//                 self.allocator.destroy(node);
//             }
//             self.head = null;
//             self.len = 0;
//         }
//
//         pub fn print(self: *This) void {
//             var current_node = self.head;
//             while (current_node) |node| {
//                 std.debug.print("centroid {any}\n", .{node.centroid});
//                 for (node.members.list.items) |member| {
//                     std.debug.print("\tmember {any}\n", .{member});
//                 }
//                 current_node = node.next;
//             }
//         }
//     };
// }
// fn asdf(centroids: std.ArrayList(@Vector(512, f32))) void {
//             var clusterIndex: usize = 0; // the index of the cluster we assign the vector to
//             var minDist: f32 = std.math.inf(f32);
//             for (centroids.items, 0..) |centroid, i| {
//                 const dist: f32 = vOps3Dimf32.dist(vec, centroid);
//                 if (dist < minDist) {
//                     minDist = dist;
//                     clusterIndex = i;
//                 }
//             }
//             try clusters.items[clusterIndex].append(vec);
//
// }
// var thread_pool: std.Thread.Pool = undefined;
//    try thread_pool.init(.{ .allocator = gpa, .n_jobs = 12 });
//    defer thread_pool.deinit();
//
//    while (true) {
//        std.debug.print("wait\n", .{});
//        var resp: std.http.Server.Response = try std.http.Server.accept(&server, .{ .allocator = gpa2 });
//        std.debug.print("conn rec\n", .{});
//        thread_pool.spawn(handleConnection, .{&resp}) catch |err| {
//            std.log.err("error spawning thread {any}", .{err});
//        };
//    }
pub fn VectorOps(comptime N: comptime_int, comptime T: type) type {
    return struct {
        pub fn dot(v1: @Vector(N, T), v2: @Vector(N, T)) T {
            return @as(T, @floatCast(@reduce(.Add, v1 * v2)));
        }

        pub fn dist(v1: @Vector(N, T), v2: @Vector(N, T)) T {
            return mag(v2 - v1);
        }

        pub fn mag(v1: @Vector(N, T)) T {
            return Q_sqrt(@as(T, @floatCast(@reduce(.Add, v1 * v1))));
        }

        pub fn cos_sim(v1: @Vector(N, T), v2: @Vector(N, T)) T {
            return dot(v1, v2) / (mag(v1) * mag(v2));
        }

        pub fn mean(group: TSA(@Vector(N, T))) @Vector(N, T) {
            var n: @Vector(N, T) = undefined;
            for (group.list.items) |point| {
                n += point;
            }

            const result: @Vector(N, T) = @splat(@floatFromInt(group.list.items.len));
            return n / result;
        }
    };
}

// test "test basic ops 3 dim" {
//     const a = @Vector(3, f32){ 1, 1, 1 };
//     const vOps3Dimf32 = VectorOps(3, f32);
//
//     const mag = vOps3Dimf32.mag(a);
//     try std.testing.expectEqual(1.7320508e0, mag);
//
//     const dist = vOps3Dimf32.dist(a, a);
//     try std.testing.expectEqual(0e0, dist);
//
//     const dot = vOps3Dimf32.dot(a, a);
//     try std.testing.expectEqual(3, dot);
//
//     const cos = vOps3Dimf32.cos_sim(a, a);
//     try std.testing.expectEqual(1, cos);
//     const test_allocator = std.testing.allocator;
//
//     var centroids = std.ArrayList(@Vector(3, f32)).init(test_allocator);
//     try centroids.append(a);
//     defer centroids.deinit();
//     const centroid = vOps3Dimf32.mean(centroids);
//     try std.testing.expectEqual(a, centroid);
// }

const threehalfs: f32 = 1.5;
pub fn Q_sqrt(number: f32) f32 {
    var i: i32 = undefined;
    var x2: f32 = undefined;
    var y: f32 = undefined;

    x2 = number * 0.5;
    y = number;
    i = @as(i32, @bitCast(y));
    i = 0x5f3759df - (i >> 1);
    y = @as(f32, @bitCast(i));
    y = y * (threehalfs - (x2 * y * y));

    return 1 / y;
}

pub fn TSA(comptime T: type) type {
    return struct {
        list: std.ArrayList(T),
        mutext: std.Thread.RwLock,

        pub fn init(allocator: *std.mem.Allocator) TSA(T) {
            return TSA(T){
                .list = std.ArrayList(T).init(allocator.*),
                .mutext = std.Thread.RwLock{},
            };
        }

        pub fn append(c: *TSA(T), value: T) !void {
            c.mutext.lock();
            defer c.mutext.unlock();
            try c.list.append(value);
        }
        pub fn getAt(c: *TSA(T), idx: usize) *T {
            c.mutext.lock();
            defer c.mutext.unlock();
            return &c.list.items[idx];
        }
        pub fn appendAt(c: *TSA(T), idx: usize, value: T) !void {
            c.mutext.lock();
            defer c.mutext.unlock();
            var a = c.list.items[idx];
            a.append(value);
        }
        pub fn deinit(c: *TSA(T)) void {
            c.mutext.lock();
            defer c.mutext.unlock();
            c.list.deinit();
        }
    };
}

// pub fn main() !void {
//     var gpa = std.heap.GeneralPurposeAllocator(.{}){};
//     var allocator = gpa.allocator();
//     defer _ = gpa.deinit();
//
//     var v = TSA(u32).init(&allocator);
//     defer v.deinit();
//
//     try v.append(13);
//     for (v.list.items) |item| {
//         std.debug.print("{any}\n", .{item});
//     }
// }
