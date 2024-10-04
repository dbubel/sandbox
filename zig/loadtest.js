import http from "k6/http";
import { sleep, check } from "k6";

export let options = {
  stages: [
    { duration: "10s", target: 10 }, // ramp up to 20 users
    // { duration: '1m', target: 20 },  // stay at 20 users
    // { duration: '30s', target: 0 },  // ramp down to 0 users
  ],
};

export default function () {
  let res = http.get("http://localhost:8080");

  check(res, {
    "status is 200": (r) => r.status === 200,
    "response time is less than 500ms": (r) => r.timings.duration < 500,
  });

  // sleep(1);
}
