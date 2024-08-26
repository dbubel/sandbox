const std = @import("std");
const kmeans = @import("kmeans_pool.zig");
const kmeans_concurrent = @import("kmeans_concurrent.zig");

pub fn main() !void {
    const allocator = std.heap.page_allocator;
    const args = try std.process.argsAlloc(allocator);

    const file = args[1];
    const number_str = args[2];
    const program = args[3];
    const number = try std.fmt.parseInt(usize, number_str, 10);

    std.debug.print("{d} {s}", .{ number, file });
    if (std.mem.eql(u8, program, "chunked")) {
        try kmeans_concurrent.run(number, file);
        return;
    }
    if (std.mem.eql(u8, program, "pooled")) {
        try kmeans.run(number, file);
        return;
    }
}
