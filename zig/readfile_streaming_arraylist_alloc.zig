const std = @import("std");
const fs = std.fs;
const print = std.debug.print;

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    const allocator = gpa.allocator();

    const file = try fs.cwd().openFile("large-file.json", .{});
    defer file.close();

    var buf_reader = std.io.bufferedReader(file.reader());
    const reader = buf_reader.reader();

    var line = std.ArrayList(u8).init(allocator);
    defer line.deinit();

    const writer = line.writer();
    while (true) {
        defer line.clearRetainingCapacity();
        reader.streamUntilDelimiter(writer, '\n', null) catch |err| {
            switch (err) {
                error.EndOfStream => {
                    return;
                },
                else => {
                    std.debug.print("{any}", .{err});
                    return;
                },
            }
        };
        // print("{s}\n", .{line.items});
    }
}
