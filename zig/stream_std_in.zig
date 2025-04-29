const std = @import("std");
const print = @import("std").debug.print;

pub fn main() !void {
    const stdin = std.io.getStdIn();
    const stdinReader = stdin.reader();
    var buf_reader = std.io.bufferedReader(stdinReader);

    var buf: [1024]u8 = undefined;

    while (true) {
        const bytesRead = buf_reader.read(&buf) catch |err| {
            print("Error: {}\n", .{err});
            return err;
        };

        if (bytesRead == 0) {
            return; // End of file
        }
    }
}
