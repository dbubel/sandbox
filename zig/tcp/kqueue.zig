const std = @import("std");
const net = std.net;
const posix = std.posix;

pub fn main() !void {
    const address = try std.net.Address.parseIp("127.0.0.1", 5882);

    const tpe: u32 = posix.SOCK.STREAM | posix.SOCK.NONBLOCK;
    const protocol = posix.IPPROTO.TCP;
    const listener = try posix.socket(address.any.family, tpe, protocol);
    defer posix.close(listener);

    try posix.setsockopt(listener, posix.SOL.SOCKET, posix.SO.REUSEADDR, &std.mem.toBytes(@as(c_int, 1)));
    try posix.bind(listener, &address.any, address.getOsSockLen());
    try posix.listen(listener, 128);

    const kfd = try posix.kqueue();
    defer posix.close(kfd);

    {
        // monitor our listening socket
        _ = try posix.kevent(kfd, &.{.{
            .ident = @intCast(listener),
            .filter = posix.system.EVFILT.READ,
            .flags = posix.system.EV.ADD,
            .fflags = 0,
            .data = 0,
            .udata = @intCast(listener),
        }}, &.{}, null);
    }

    var ready_list: [128]posix.Kevent = undefined;
    while (true) {
        const ready_count = try posix.kevent(kfd, &.{}, &ready_list, null);
        for (ready_list[0..ready_count]) |ready| {
            const ready_socket: i32 = @intCast(ready.udata);
            if (ready_socket == listener) {
                const client_socket = try posix.accept(listener, null, null, posix.SOCK.NONBLOCK);
                errdefer posix.close(client_socket);
                _ = try posix.kevent(kfd, &.{.{
                    .ident = @intCast(client_socket),
                    .flags = posix.system.EV.ADD,
                    .filter = posix.system.EVFILT.READ,
                    .fflags = 0,
                    .data = 0,
                    .udata = @intCast(client_socket),
                }}, &.{}, null);
            } else {
                var closed = false;
                var buf: [4096]u8 = undefined;
                const read = posix.read(ready_socket, &buf) catch 0;
                if (read == 0) {
                    closed = true;
                } else {
                    std.debug.print("[{d}] got: {any}\n", .{ ready_socket, buf[0..read] });
                }

                if (closed) {
                    posix.close(ready_socket);
                }
            }
        }
    }
}
