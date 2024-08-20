const std = @import("std");
pub fn main() !void {
    const file = try std.fs.cwd().openFile("output.txt", .{});
    defer file.close();

    var buf_reader = std.io.bufferedReader(file.reader());
    const reader = buf_reader.reader();

    var line: [512]u8 = undefined;
    var writer = std.io.fixedBufferStream(&line);

    while (true) {
        defer writer.reset();
        reader.streamUntilDelimiter(writer.writer(), '\n', null) catch |err| {
            switch (err) {
                error.EndOfStream => {
                    return;
                },
                else => {
                    std.debug.print("{any}", .{err});
                },
            }
        };
    }
}
