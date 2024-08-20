const std = @import("std");
const fixedBufferStream = @import("std").io.fixedBufferStream;
const print = @import("std").debug.print;

pub fn main() !void {
    const stdin = std.io.getStdIn();
    const stdinReader = stdin.reader();

    var buf: [1024]u8 = undefined;
    var writer = fixedBufferStream(&buf);

    while (true) {
        defer writer.reset();
        stdinReader.streamUntilDelimiter(writer.writer(), '\n', null) catch |err| {
            switch (err) {
                error.EndOfStream => {
                    return;
                },
                else => {
                    print("err", .{});
                    return;
                },
            }
        };
        print("{s}\n", .{buf[0..writer.pos]});
    }
}
