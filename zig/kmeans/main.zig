const std = @import("std");
const kmeans = @import("kmeans.zig");

pub fn main() !void {
    const allocator = std.heap.page_allocator;
    var args = try std.process.ArgIterator.initWithAllocator(allocator);

    // const stdout = std.io.getStdOut().writer();
    var asdf: []const u8 = undefined;
    while (args.next()) |arg| {
        // try stdout.print("{s}\n", .{arg});
        asdf = arg;
    }
    const x = try std.fmt.parseInt(usize, asdf, 10);
    try kmeans.run(x);
}
