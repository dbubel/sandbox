const std = @import("std");
const t = @import("thread_safe_arraylist.zig");

pub fn main() !void {
    var wg = std.Thread.WaitGroup{};
    var gpa = std.heap.GeneralPurposeAllocator(.{ .thread_safe = true }){};
    var allocator = gpa.allocator();
    defer _ = gpa.deinit();

    // var arrayList = t.ThreadSafeArrayList(@Vector(2, f32)).init(&allocator);
    var arrayList = t.ThreadSafeArrayList(t.ThreadSafeArrayList(@Vector(2, f32))).init(&allocator);
    defer {
        // Deinitialize each array list inside the clusters array list
        for (arrayList.list.items) |*array_item| {
            array_item.deinit();
        }
        defer arrayList.deinit();
    }

    for (0..10) |_| {
        arrayList.append(t.ThreadSafeArrayList(@Vector(2, f32)).init(&allocator)) catch unreachable;
    }

    var thread_pool: std.Thread.Pool = undefined;
    try thread_pool.init(.{ .allocator = allocator, .n_jobs = 12 });
    defer thread_pool.deinit();

    for (0..5) |_| {
        for (0..10) |i| {
            thread_pool.spawnWg(&wg, f, .{ &arrayList, i });
        }
    }

    thread_pool.waitAndWork(&wg);

    for (arrayList.list.items) |item| {
        std.debug.print("value: {any}\n", .{item});
    }
    std.debug.print("done\n", .{});
}

fn f(al: *t.ThreadSafeArrayList(t.ThreadSafeArrayList(@Vector(2, f32))), i: usize) void {
    std.debug.print("started...\n", .{});
    std.time.sleep(std.time.ns_per_s * 0.5);
    var a = al.getAt(i);
    a.append(@Vector(2, f32){ 1, 2 }) catch |err| {
        std.debug.print("{any}\n", .{err});
    };
}
