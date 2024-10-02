const std = @import("std");

pub fn main() !void {
    const result = try Calculator.init(10)
        .add(5)
        .multiply(2)
        .subtract(3)
        .divide(2)
        .result();

    std.debug.print("Result: {d}\n", .{result});
}

const Calculator = struct {
    value: f64,

    pub fn init(initial: f64) !Calculator {
        return Calculator{ .value = initial };
    }

    pub fn add(self: Calculator, x: f64) Calculator {
        return .{ .value = self.value + x };
    }

    pub fn subtract(self: Calculator, x: f64) Calculator {
        return .{ .value = self.value - x };
    }

    pub fn multiply(self: Calculator, x: f64) Calculator {
        return .{ .value = self.value * x };
    }

    pub fn divide(self: Calculator, x: f64) Calculator {
        return .{ .value = self.value / x };
    }

    pub fn result(self: Calculator) f64 {
        return self.value;
    }
};
