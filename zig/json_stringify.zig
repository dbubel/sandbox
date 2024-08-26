const std = @import("std");
const jsonStringify = @import("std").json.stringify;

fn marshal(T: anytype) !void {
    const out = std.io.getStdOut().writer();
    try jsonStringify(T, .{}, out);
}

test "marshal" {
    const data = .{
        .name = "Ziguana",
        .id = 1234,
        .is_active = true,
    };
    try marshal(data);
}
