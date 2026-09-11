import { createRouter } from "@solidjs/router";

import ArenaRoute from "./routes/ArenaRoute";
import GameRoute from "./routes/GameRoute";
import HomeRoute from "./routes/HomeRoute";
import NotFoundRoute from "./routes/NotFoundRoute";
import ProfileRoute from "./routes/ProfileRoute";
import ProofOfPlayRoute from "./routes/ProofOfPlayRoute";
import SquadRoute from "./routes/SquadRoute";

export const AppRouter = createRouter({
  routes: [
    { path: "/", component: HomeRoute },
    { path: "/game", component: GameRoute },
    { path: "/profile", component: ProfileRoute },
    { path: "/arena", component: ArenaRoute },
    { path: "/squad", component: SquadRoute },
    { path: "/proof-of-play", component: ProofOfPlayRoute },
    { path: "*all", component: NotFoundRoute },
  ],
});
