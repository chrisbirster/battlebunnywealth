import { createRouter } from "@solidjs/router";

import GameRoute from "./routes/GameRoute";
import HomeRoute from "./routes/HomeRoute";
import NotFoundRoute from "./routes/NotFoundRoute";
import ProofOfPlayRoute from "./routes/ProofOfPlayRoute";
import SquadRoute from "./routes/SquadRoute";

export const AppRouter = createRouter({
  routes: [
    { path: "/", component: HomeRoute },
    { path: "/game", component: GameRoute },
    { path: "/squad", component: SquadRoute },
    { path: "/proof-of-play", component: ProofOfPlayRoute },
    { path: "*all", component: NotFoundRoute },
  ],
});
