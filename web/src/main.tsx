import { render } from "@solidjs/web";

import "./index.css";
import { AppRouter } from "./router";

if (import.meta.env.DEV) {
  void import("virtual:stylex:runtime");
  void import("virtual:stylex:css-only");
}

render(() => <AppRouter />, document.getElementById("root") as HTMLElement);
