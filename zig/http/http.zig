const std = @import("std");

const Server = struct {
    port: u16,
    pub fn init(port: u16) Server {
        return Server{ .port = port };
    }

    pub fn run(s: Server) !void {
        const addr = std.net.Address.parseIp4("127.0.0.1", s.port) catch |err| {
            std.debug.print("{any}\n", .{err});
            return;
        };

        var base_server = addr.listen(.{ .kernel_backlog = 1024, .reuse_address = true }) catch |err| {
            std.debug.print("error listen {any}\n", .{err});
            return;
        };

        defer base_server.deinit();

        var gpa = std.heap.GeneralPurposeAllocator(.{}){};
        defer _ = gpa.deinit();

        const num_threads = 12; //try std.Thread.getCpuCount();
        const threads = try gpa.allocator().alloc(std.Thread, num_threads);

        for (0.., threads) |i, *t| {
            t.* = try std.Thread.spawn(.{}, handlerThread, .{ &base_server, i });
        }

        std.debug.print("starting...\n", .{});
        for (threads) |t| {
            t.join();
        }
    }
};

fn handlerThread(base_server: *std.net.Server, i: usize) void {
    _ = i;
    var header_buf: [1024]u8 = undefined;
    while (true) {
        var conn = base_server.accept() catch |err| {
            std.debug.print("error accept {any}\n", .{err});
            return;
        };
        defer conn.stream.close();

        var server = std.http.Server.init(conn, &header_buf);
        var req = server.receiveHead() catch |err| {
            std.debug.print("{any}\n", .{err});
            return;
        };
        const reader = req.reader() catch unreachable;
        const n = reader.readAll(&header_buf) catch unreachable;

        _ = n;
        // _ = conn.stream.writeAll("hello fuck you") catch unreachable;
        var aids = req.respondStreaming(.{ .send_buffer = &header_buf });
        // _ = aids;

        _ = std.json.stringify(.{ .fuck = "fuck" }, .{ .whitespace = .minified }, aids.writer()) catch unreachable;
        _ = aids.flush() catch unreachable;
        _ = aids.end() catch unreachable;
        //
        // _ = req.respond("Hello http!\n", .{ .status = .ok }) catch unreachable;

        // std.debug.print("   Handling request for {s}\n", .{req.head.target});
        // std.debug.print("   bytes read {d} {s}\n", .{ n, header_buf[0..n] });
    }
}

pub fn main() !void {
    const server = Server.init(8080);
    try server.run();
}
