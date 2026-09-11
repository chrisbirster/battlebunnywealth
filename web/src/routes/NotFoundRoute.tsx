import * as stylex from "@stylexjs/stylex";
import Shell from "../components/Shell";

export default function NotFoundRoute() {
  return <Shell><section {...stylex.props(styles.wrap)}><div {...stylex.props(styles.code)}>404</div><h1>Wrong burrow.</h1><p>This tunnel does not go anywhere useful.</p><a href="/">Return to command</a></section></Shell>;
}

const styles = stylex.create({
  wrap: { maxWidth: "900px", margin: "0 auto", padding: "100px 24px" },
  code: { color: "#d9f13b", fontWeight: 900, letterSpacing: "0.16em" },
});
