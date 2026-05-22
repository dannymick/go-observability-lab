import http from "k6/http";
import { sleep, check } from "k6";

export const options = {
  vus: 10,
  duration: "30s",
};

export default function () {
  const res = http.get("http://localhost:8080/work");

  check(res, {
    "status is 200 or 500": (r) => r.status === 200 || r.status === 500,
  });

  sleep(0.2);
}