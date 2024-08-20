const std = @import("std");
const rand = std.crypto.random;
const DIMS = 512; // dimension of the vectors we are working with
const VType = @Vector(DIMS, f32);

const CLUSTERS: usize = 3; // number of clusters to build
const EPSILON: f32 = 0.1;
const THREADS = 12;
const vecOps = VectorOps(DIMS, f32);

pub fn main() !void {
    const cpuCount = try std.Thread.getCpuCount();
    const allocator = std.heap.c_allocator;

    var threadBuckets: [THREADS][CLUSTERS](std.ArrayList(VType)) = undefined;
    for (0..THREADS) |i| {
        for (0..CLUSTERS) |k| {
            threadBuckets[i][k] = std.ArrayList(VType).init(allocator);
        }
    }

    defer {
        for (0..THREADS) |i| {
            for (0..CLUSTERS) |k| {
                threadBuckets[i][k].deinit();
            }
        }
    }
    // _ = threadBuckets;

    var thread_pool: std.Thread.Pool = undefined;
    try thread_pool.init(.{ .allocator = allocator, .n_jobs = cpuCount });
    defer thread_pool.deinit();

    var buf: [1024 * 1024]u8 = undefined;
    var writer = std.io.fixedBufferStream(&buf);

    var centroids = std.ArrayList(@Vector(DIMS, f32)).init(allocator);
    defer centroids.deinit();

    // vector data from file
    // const allocator = std.heap.page_allocator;

    // Allocate the large array on the heap
    var vecData: []VType = try allocator.alloc(@Vector(DIMS, f32), 20000);

    // if (true) {
    //     return;
    // }
    // ----------------------------------------------------------------- File read start
    var t = std.time.milliTimestamp();
    const file = try std.fs.cwd().openFile("../data/512_10k.jsonl", .{});
    var buffered = std.io.bufferedReader(file.reader());
    var reader = buffered.reader();
    defer file.close();

    var numVecs: usize = 0;
    while (true) {
        defer writer.reset();
        // read the file line by line writing into buf
        reader.streamUntilDelimiter(writer.writer(), '\n', null) catch |err| {
            switch (err) {
                error.EndOfStream => {
                    break;
                },
                else => {
                    std.debug.print("{any}", .{err});
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
                    std.debug.print("{any} error reading json\n", .{err});
                    break;
                },
            }
        };

        vecData[numVecs] = arr.value;
        numVecs += 1;
        // try vecData.append(arr.value);
        defer arr.deinit();
    }
    std.debug.print("File read time: {d}ms\n", .{std.time.milliTimestamp() - t});
    // ----------------------------------------------------------------- File read end

    // pick random points to use as centroids
    // TODO change this to a map so we dont pick the same points
    t = std.time.milliTimestamp();
    var goodClusters: usize = 0;
    while (goodClusters < CLUSTERS) {
        const d = rand.intRangeAtMost(usize, 0, numVecs - 1);
        var exists = false;
        for (centroids.items) |existing| {
            if (vecOps.equal(existing, vecData[d])) {
                exists = true;
                break;
            }
        }
        if (exists) {
            continue;
        }
        try centroids.append(vecData[d]);
        goodClusters += 1;
    }
    // try centroids.append(VType{ 3, 2 });
    // try centroids.append(VType{ 350, 420 });

    // std.debug.print("Total vectors: {d}\n", .{numVecs});
    // for (vecData[0..numVecs]) |i| {
    //     std.debug.print("{any}\n", .{i});
    // }
    //
    for (centroids.items) |i| {
        std.debug.print("Centroids:\n", .{});
        std.debug.print("\t{d}\n", .{i});
    }
    // std.debug.print("\n", .{});
    var wg = std.Thread.WaitGroup{};
    const inc: usize = (numVecs + cpuCount) / cpuCount;
    // std.debug.print("Increment: {d}\n", .{inc});

    // This code iterates over vecData in chunks determined by inc,
    // ensuring that each chunk is processed without exceeding the array boundaries.
    var clusters = [_]std.ArrayList(VType){std.ArrayList(VType).init(allocator)} ** CLUSTERS;
    t = std.time.milliTimestamp();
    // while (true) {
    for (0..10) |round| {
        defer std.debug.print("round {d}\n", .{round});
        var i: usize = 0;
        var loop_num: usize = 0;
        while (i < numVecs) : (i += inc) {
            defer loop_num += 1;
            wg.start();
            // Calculate the end index, ensuring it doesn't exceed the array length
            const end = if (i + inc < numVecs) i + inc else numVecs;
            try thread_pool.spawn(calulateAndAssign, .{ &wg, vecData[i..end], &threadBuckets[loop_num], centroids });
        }

        thread_pool.waitAndWork(&wg);
        wg.reset();
        // var newClusters =
        threadBuckets = threadBuckets;

        for (0..THREADS) |j| {
            for (0..CLUSTERS) |k| {
                if (threadBuckets[j][k].items.len > 0) {
                    // TODO dont loop here just mem copy or something
                    for (threadBuckets[j][k].items) |item| {
                        try clusters[k].append(item);
                    }
                }
            }
        }

        // // for (clusters) |cluster| {
        // //     std.debug.print("cluster {any}\n", .{cluster.items});
        // // }
        //
        var moved: bool = false;
        for (centroids.items, clusters) |*centroid, cluster| {
            const mean = vecOps.mean(cluster);
            const new_centroid = mean;
            const dist = vecOps.dist(centroid.*, new_centroid);
            // std.debug.print("dist: {d} old centroid: {any} new centroid: {any}\n", .{ dist, centroid.*, new_centroid });
            // std.debug.print("centroid: {d} mean: {d} dist: {d}\n", .{ centroid.*, mean, dist });
            std.debug.print("dist: {d}\n", .{dist});
            if (dist > EPSILON) {
                moved = true;
            }

            centroid.* = new_centroid;
        }
        if (!moved) {
            std.debug.print("done\n", .{});
            break;
        }
        //
        for (0..THREADS) |x| {
            for (0..CLUSTERS) |y| {
                threadBuckets[x][y].clearRetainingCapacity();
            }
        }
        for (&clusters) |*c| {
            c.clearRetainingCapacity();
        }
    }
    // for (clusters) |cluster| {
    //     std.debug.print("\n", .{});
    //     for (cluster.items) |item| {
    //         std.debug.print("{any}\n", .{item});
    //     }
    // }
    std.debug.print("kmeans time: {any}\n", .{std.time.milliTimestamp() - t});
    // std.debug.print("moved {any}\n", .{moved});
}

fn calulateAndAssign(wg: *std.Thread.WaitGroup, chunk: []VType, cluster: []std.ArrayList(VType), centroids: std.ArrayList(VType)) void {
    defer wg.finish();
    var bestCluster: usize = 0; // the index of the cluster we assign the vector to
    var minDist: f32 = std.math.inf(f32);
    // var bestCentroid: VType = undefined;
    // std.debug.print("centroid in thread {any}\n", .{centroids.items});

    for (chunk) |vec| {
        for (centroids.items, 0..) |centroid, i| {
            const dist: f32 = vecOps.dist(vec, centroid);
            if (dist < minDist) {
                minDist = dist;
                bestCluster = i;
                // bestCentroid = centroid;
            }
        }
        cluster[bestCluster].append(vec) catch |err| {
            std.debug.print("error {any}\n", .{err});
        };
    }

    // for (cluster, 0..) |_, i| {
    //     if (cluster[i].items.len > 0) {
    //         std.debug.print("idx :{d} vecs: {any}\n", .{ i, cluster[i].items });
    //     }
    // }
}

pub fn VectorOps(comptime N: comptime_int, comptime T: type) type {
    return struct {
        pub fn dot(v1: @Vector(N, T), v2: @Vector(N, T)) T {
            return @as(T, @floatCast(@reduce(.Add, v1 * v2)));
        }

        pub fn dist(v1: @Vector(N, T), v2: @Vector(N, T)) T {
            return mag(v2 - v1);
        }

        pub fn mag(v1: @Vector(N, T)) T {
            return std.math.sqrt(@as(T, @floatCast(@reduce(.Add, v1 * v1))));
        }

        pub fn cos_sim(v1: @Vector(N, T), v2: @Vector(N, T)) T {
            return dot(v1, v2) / (mag(v1) * mag(v2));
        }

        pub fn mean(group: std.ArrayList(@Vector(N, T))) @Vector(N, T) {
            var n: @Vector(N, T) = undefined;
            for (group.items) |point| {
                n += point;
            }

            const result: @Vector(N, T) = @splat(@floatFromInt(group.items.len));
            return n / result;
        }
        pub fn equal(v1: @Vector(N, T), v2: @Vector(N, T)) bool {
            for (0..N) |i| {
                if (v1[i] != v2[i]) {
                    return false;
                }
            }
            return true;
        }
    };
}
test "vec mean" {
    const alloc = std.testing.allocator;
    const asdf = VectorOps(2, f32);
    var arr = std.ArrayList(@Vector(2, f32)).init(alloc);
    defer arr.deinit();

    try arr.append(@Vector(2, f32){ 1, 1 });
    try arr.append(@Vector(2, f32){ 2, 2 });
    const mean = asdf.mean(arr);
    std.debug.print("mean: {d}\n", .{mean});
}
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
