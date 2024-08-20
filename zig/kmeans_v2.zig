const std = @import("std");
const rand = std.crypto.random;
const DIMS = 8; // dimension of the vectors we are working with
const VType = @Vector(DIMS, f32);

const K: usize = 10; // number of clusters to build
const EPSILON: f32 = 0.1;
const THREADS = 12;
const vecOps = VectorOps(DIMS, f32);

const Vector = struct {
    vec: VType,
    centroid: VType,
    cluster_id: usize,
};

pub fn main() !void {
    var wg = std.Thread.WaitGroup{};
    const cpuCount = try std.Thread.getCpuCount();
    const allocator = std.heap.c_allocator;

    var thread_pool: std.Thread.Pool = undefined;
    try thread_pool.init(.{ .allocator = allocator, .n_jobs = cpuCount });
    defer thread_pool.deinit();

    var buf: [1024 * 10]u8 = undefined;
    var writer = std.io.fixedBufferStream(&buf);

    var centroids = std.ArrayList(@Vector(DIMS, f32)).init(allocator);
    defer centroids.deinit();

    var vecData: []VType = try allocator.alloc(@Vector(DIMS, f32), 1000000);

    // ----------------------------------------------------------------- File read start
    var t = std.time.milliTimestamp();
    const file = try std.fs.cwd().openFile("../data/8_200k.jsonl", .{});
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
    // for (centroids.items) |centroid| {
    //     std.debug.print("Centroid {d}\n", .{centroid});
    // }

    const inc: usize = (numVecs + cpuCount) / cpuCount;
    // var clusters = std.ArrayList(Cluster).init(allocator);
    // for (0..K, centroids.items) |_, centroid| {
    //     try clusters.append(Cluster.init(allocator, centroid));
    // }
    std.debug.print("chunk size {d}\n", .{inc});

    // t = std.time.milliTimestamp();
    // while (true) {
    var clusters = std.ArrayList(std.ArrayList(Vector)).init(allocator);
    for (0..THREADS) |_| {
        var a = std.ArrayList(Vector).init(allocator);
        try a.ensureTotalCapacity(1000000);
        try clusters.append(a);
    }
    var cluster_sum = std.ArrayList(std.ArrayList(Vector)).init(allocator);
    for (0..K) |_| {
        var a = std.ArrayList(Vector).init(allocator);
        try a.ensureTotalCapacity(1000000);
        try cluster_sum.append(a);
    }

    t = std.time.milliTimestamp();
    var total_iterations: usize = 0;
    while (true) {
        total_iterations += 1;
        var i: usize = 0;
        var loop_num: usize = 0;
        // t = std.time.milliTimestamp();
        while (i < numVecs) : (i += inc) {
            defer loop_num += 1;
            const end = if (i + inc < numVecs) i + inc else numVecs;

            wg.start();
            try thread_pool.spawn(calulateAndAssign, .{ &wg, vecData[i..end], centroids, &clusters.items[loop_num] });
        }

        thread_pool.waitAndWork(&wg);
        wg.reset();
        // std.debug.print("thread time: {d}\n", .{std.time.milliTimestamp() - t});
        for (clusters.items) |item| {
            for (item.items) |v| {
                try cluster_sum.items[v.cluster_id].append(v);
                // std.debug.print("{d}", .{v.cluster_id});
            }
        }
        var moved: bool = false;
        // t = std.time.milliTimestamp();
        for (cluster_sum.items, centroids.items) |sum, *centroid| {
            // _ = centroid;
            const new_centroid = vecOps.meanV(sum);
            const dist = vecOps.dist(centroid.*, new_centroid);
            if (dist > EPSILON) {
                moved = true;
            }
            // for (sum.items) |s| {
            //     std.debug.print("{any}\n", .{s});
            // }
            // std.debug.print("dist: {d} new_centroid {d} old_centroid {d}\n\n", .{ dist, new_centroid, centroid.* });

            centroid.* = new_centroid;
        }
        // std.debug.print("move centroids time: {d}\n", .{std.time.milliTimestamp() - t});
        if (!moved) {
            std.debug.print("done\n", .{});
            break;
        }
        for (clusters.items) |*c| {
            c.clearRetainingCapacity();
        }

        for (cluster_sum.items) |*c| {
            c.clearRetainingCapacity();
        }
    }
    std.debug.print("total iterations {d}\n", .{total_iterations});
    std.debug.print("File read time: {d}ms\n", .{std.time.milliTimestamp() - t});
}

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
        // std.debug.print("vec: {d} centroid: {d} dist: {d}\n", .{ vec, bestCentroid, minDist });
    }
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
            // return Q_sqrt(@as(T, @floatCast(@reduce(.Add, v1 * v1))));
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
            return @as(T, @floatCast(@reduce(.Add, (v2 - v1) * (v2 - v1))));
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
