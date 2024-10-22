const std = @import("std");
const rand = std.crypto.random;
const DIMS = 2; // dimension of the vectors we are working with
const VType = @Vector(DIMS, f32);

const EPSILON: f32 = 0.01;
const THREADS = 12;
const vecOps = VectorOps(DIMS, f32);

const Vector = struct {
    vec: VType,
    centroid: VType,
    cluster_id: usize,
};

pub fn run(K: usize, file_name: []const u8) !void {
    const cpuCount = try std.Thread.getCpuCount();
    const allocator = std.heap.c_allocator;

    var thread_pool: std.Thread.Pool = undefined;
    try thread_pool.init(.{ .allocator = allocator, .n_jobs = cpuCount });
    defer thread_pool.deinit();

    var centroids = std.ArrayList(@Vector(DIMS, f32)).init(allocator);
    defer centroids.deinit();

    var vecData: []VType = try allocator.alloc(@Vector(DIMS, f32), 1100000);

    // ----------------------------------------------------------------- File read start
    var buf: [1024 * 256]u8 = undefined;
    var writer = std.io.fixedBufferStream(&buf);
    const file = try std.fs.cwd().openFile(file_name, .{});
    var buffered = std.io.bufferedReader(file.reader());
    var reader = buffered.reader();
    defer file.close();
    var numVecs: usize = 0;

    var t = std.time.milliTimestamp();
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
            std.debug.print("error: {any}\n", .{err});
            break;
        };

        vecData[numVecs] = arr.value;
        numVecs += 1;
        defer arr.deinit();
    }
    std.debug.print("File read time: {d}ms\n", .{std.time.milliTimestamp() - t});
    // ----------------------------------------------------------------- File read end

    // pick random points to use as centroids
    var goodClusters: usize = 0;
    while (goodClusters < K) {
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

    const inc: usize = (numVecs + cpuCount) / cpuCount;
    var clusters = std.ArrayList(std.ArrayList(Vector)).init(allocator);
    for (0..THREADS) |_| {
        const a = std.ArrayList(Vector).init(allocator);
        try clusters.append(a);
    }

    var cluster_sum = std.ArrayList(std.ArrayList(Vector)).init(allocator);
    for (0..K) |_| {
        const a = std.ArrayList(Vector).init(allocator);
        try cluster_sum.append(a);
    }

    var total_iterations: usize = 0;
    t = std.time.milliTimestamp();
    var wg = std.Thread.WaitGroup{};
    while (true) {
        total_iterations += 1;
        var i: usize = 0;
        var loop_num: usize = 0;
        while (i < numVecs) : (i += inc) {
            defer loop_num += 1;
            const end = if (i + inc < numVecs) i + inc else numVecs;

            wg.start();
            try thread_pool.spawn(calulateAndAssign, .{ &wg, vecData[i..end], centroids, &clusters.items[loop_num] });
        }

        thread_pool.waitAndWork(&wg);
        wg.reset();

        for (clusters.items) |item| {
            for (item.items) |v| {
                try cluster_sum.items[v.cluster_id].append(v);
            }
        }

        var moved: bool = false;
        for (cluster_sum.items, centroids.items) |sum, *centroid| {
            const new_centroid = vecOps.meanV(sum);
            const dist = vecOps.dist(centroid.*, new_centroid);
            if (dist > EPSILON) {
                moved = true;
            }
            centroid.* = new_centroid;
        }

        if (!moved) {
            break;
        }

        for (clusters.items) |*c| {
            c.clearRetainingCapacity();
        }

        for (cluster_sum.items) |*c| {
            c.clearRetainingCapacity();
        }
    }
    for (clusters.items) |item| {
        std.debug.print("\n", .{});
        for (item.items) |a| {
            std.debug.print("{any}\n", .{a});
        }
    }

    std.debug.print("total iterations {d}\n", .{total_iterations});
    std.debug.print("kmeans time: {d}ms\n", .{std.time.milliTimestamp() - t});
}

// fn marshal(T: anytype) !void {
//     const out = std.io.getStdOut().writer();
//     try std.json.stringify(T, .{}, out);
//     _ = out.write("\n") catch |err| {
//         std.debug.print("{any}", .{err});
//     };
// }

fn calulateAndAssign(wg: *std.Thread.WaitGroup, chunk: []VType, centroids: std.ArrayList(VType), classifiedVecs: *std.ArrayList(Vector)) void {
    defer wg.finish();

    for (chunk) |vec| {
        var bestCluster: usize = 0; // the index of the cluster we assign the vector to
        var minDist: f32 = std.math.inf(f32);
        var bestCentroid: VType = undefined;

        for (centroids.items, 0..) |centroid, i| {
            const dist: f32 = vecOps.distSquared(vec, centroid);
            if (dist < minDist) {
                minDist = dist;
                bestCluster = i;
                bestCentroid = centroid;
            }
        }
        const v = Vector{ .cluster_id = bestCluster, .vec = vec, .centroid = bestCentroid };
        classifiedVecs.append(v) catch |err| {
            std.debug.print("err {any}\n", .{err});
        };
    }
}

pub fn VectorOps(comptime N: comptime_int, comptime T: type) type {
    return struct {
        pub fn dot(v1: @Vector(N, T), v2: @Vector(N, T)) T {
            return @as(T, @reduce(.Add, v1 * v2));
        }

        pub fn dist(v1: @Vector(N, T), v2: @Vector(N, T)) T {
            return mag(v2 - v1);
        }

        pub fn mag(v1: @Vector(N, T)) T {
            return @sqrt(@as(T, @floatCast(@reduce(.Add, v1 * v1))));
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

        pub fn distSquared(v1: @Vector(N, T), v2: @Vector(N, T)) T {
            return @as(T, @reduce(.Add, (v2 - v1) * (v2 - v1)));
        }

        pub fn meanV(group: std.ArrayList(Vector)) @Vector(N, T) {
            var n: @Vector(N, T) = undefined;
            for (group.items) |point| {
                n += point.vec;
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
