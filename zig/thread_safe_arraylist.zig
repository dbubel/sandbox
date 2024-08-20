const std = @import("std");

pub fn ThreadSafeArrayList(comptime T: type) type {
    return struct {
        list: std.ArrayList(T),
        mutext: std.Thread.Mutex,

        pub fn init(allocator: *std.mem.Allocator) ThreadSafeArrayList(T) {
            return ThreadSafeArrayList(T){
                .list = std.ArrayList(T).init(allocator.*),
                .mutext = std.Thread.Mutex{},
            };
        }

        pub fn append(c: *ThreadSafeArrayList(T), value: T) !void {
            c.mutext.lock();
            defer c.mutext.unlock();
            try c.list.append(value);
        }
        pub fn getAt(c: *ThreadSafeArrayList(T), idx: usize) *T {
            return &c.list.items[idx];
        }
        pub fn appendAt(c: *ThreadSafeArrayList(T), idx: usize, value: T) !void {
            c.list.items[idx].append(value);
        }
        pub fn deinit(c: *ThreadSafeArrayList(T)) void {
            c.mutext.lock();
            defer c.mutext.unlock();
            c.list.deinit();
        }
    };
}

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    var allocator = gpa.allocator();
    defer _ = gpa.deinit();

    var v = ThreadSafeArrayList(u32).init(&allocator);
    defer v.deinit();

    try v.append(13);
    for (v.list.items) |item| {
        std.debug.print("{any}\n", .{item});
    }
}
