const std = @import("std");
const kmeans_concurrent = @import("kmeans_concurrent.zig");

pub fn main() !void {
    const allocator = std.heap.page_allocator;
    const args = try std.process.argsAlloc(allocator);

    const file = args[1];
    const number_str = args[2];
    const num_clusters = try std.fmt.parseInt(usize, number_str, 10);

    std.debug.print("{s} {d}\n", .{ file, num_clusters });
    try kmeans_concurrent.run(num_clusters, file);
}
