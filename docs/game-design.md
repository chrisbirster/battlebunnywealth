# Game design

## Fantasy

Battle Bunny Wealth is a military-comedy world populated by rabbits who are chronically underfunded, wildly overconfident, and obsessed with getting rich. The player creates their own Battle Bunny, joins the Warren, builds businesses, competes through seasons, and eventually fights other players in short grid-bomb arena matches.

The tone is cartoon military parody, not realistic warfare. Explosions are arcade tools. The world should feel like an absurd military unit crossed with an idle business empire.

## Core product principles

1. **Idle progress feels good.** Open the app, collect meaningful progress, make a few decisions, leave.
2. **Every ranked arena match starts equal.** Wealth, CARROT balance, account age, purchases, NPC progression, and Proof-of-Play authority cannot buy combat power.
3. **Warren Wars is skill based.** Movement, timing, prediction, trapping, map control, and adaptation determine winners.
4. **NPCs belong to the world; the player's bunny belongs to the player.** Named NPCs drive story, tutorials, missions, jokes, and lore. Players create and name their own bunny.
5. **Network participation is optional for enjoying the game.** Missions can increase Proof-of-Play authority, but a player can build businesses and play Warren Wars without being forced to perform network missions.
6. **The blockchain is invisible until useful.** Nobody should need to understand epochs, attestations, or committees to enjoy Battle Bunny Wealth.
7. **You can buy style. You can earn status. You cannot buy skill.**

## The three games

Battle Bunny Wealth intentionally combines three related but distinct competitions:

```text
ECONOMIC GAME
How efficiently can I build my seasonal business empire?
        |
        v
season wealth / standings / cosmetic and collection rewards

NETWORK GAME
How consistently and honestly do I participate in optional missions?
        |
        v
Proof-of-Play authority / validator eligibility / CARROT eligibility

SKILL GAME
How good am I at Warren Wars?
        |
        v
wins / rating / championships / competitive reputation
```

No one axis should automatically dominate the other two.

## Player identity

The player is the protagonist. During onboarding the player creates their own Battle Bunny.

Initial customization space:

- bunny name
- callsign
- fur color/pattern
- ears
- face/scars
- uniform
- helmet/headgear
- glasses
- patches
- backpack
- dog tags
- bomb skin
- victory animation
- profile banner/title

Customization is cosmetic in ranked Warren Wars.

A profile can eventually expose things such as:

```text
Name: Sergeant Fluffington
Callsign: Boomstick
Rank: Corporal
Warren: The Dusty Burrow
Season Wealth: $84.2M
Warren Wars: 31-12
Proof-of-Play Authority: 428
CARROT: 17.4
```

Military rank is earned through progression rather than freely selected. The final fictional rank ladder does not need to copy a real military rank structure exactly.

## NPC cast

Named characters are **not the player's avatar and are not the default playable competitive characters**. They are story/lore NPCs who make the Warren feel alive.

- **First Sergeant Hard-as-Nails** — terrifying senior enlisted mentor and mission authority.
- **Da Champ** — famous arena personality/rival who embodies competitive Warren Wars.
- **Private Stuffy** — rookie, comic relief, and early-game companion.
- **Corporal Boomboom** — demolition specialist who introduces bombs and arena concepts.
- **Captain Cashmere** — logistics/finance officer who introduces businesses and optimization.
- **Doc Flopsy** — medic/support NPC and exhausted professional surrounded by bad decisions.

NPCs can give missions, introduce mechanics, appear in seasonal stories, operate businesses in the fiction, and react to player progress. Their lore does not grant a player's ranked combat statistics.

## Seasonal loop

The idle game is structured around seasons.

```text
SEASON START
    |
build and upgrade bunny businesses
    |
maximize passive seasonal income
    |
optional missions increase Proof-of-Play authority
    |
season standings / story progression
    |
turn in seasonal wealth for persistent rewards/items
    |
Warren Wars competition continues with equal-start rules
    |
NEXT SEASON
```

The important reset boundary is that seasonal business wealth is a game resource, not CARROT and not permanent combat power.

The exact season length is TBD and should be tuned from playtesting rather than chosen for token-economy reasons.

## Idle business layer

Players build and optimize businesses operated by their warren. Candidate operations include:

- Questionable Carrot Logistics
- Scrap Recovery Detail
- Tactical Lawn Services
- Surplus Snack Depot
- Warren Energy Drink Company
- Burrow Freight & Salvage
- Officer's Club Gift Shop
- Demolition Contracting
- Tunnel Toll Authority
- Bunker Casino

Each business can have levels, managers/NPC story beats, production rates, upgrade choices, synergies, offline earnings, and seasonal objectives.

The goal is to create the satisfying optimization loop of an idle tycoon game: players compare strategies and try to maximize the money their businesses generate before the season closes.

## Season turn-in

At season end, seasonal wealth can be exchanged for items and permanent collection/progression rewards. These rewards must respect competitive equality.

Good reward categories:

- bunny cosmetics
- uniforms/headgear
- bomb skins
- explosion visual effects
- victory poses/emotes
- profile banners/titles
- medals/trophies
- base/warren decorations
- collectible blueprints that change presentation rather than ranked power
- story unlocks
- cosmetic arena themes

A season reward must not let a wealthy player enter ranked Warren Wars with extra speed, a larger blast radius, more starting bombs, more health, or any equivalent combat advantage.

## Warren Wars

Warren Wars is an original grid-bomb arena inspired by the readability, mind games, and tension of classic maze bomb games. It must use original characters, art, maps, sounds, presentation, and implementation.

### Ranked equality rule

Every ranked match begins from a standardized competitive state.

Example baseline:

```text
movement speed: equal
starting bombs: equal
starting blast radius: equal
health/lives: equal
starting abilities: equal by ruleset
wallet/CARROT effect: none
season wealth effect: none
Proof-of-Play authority effect: none
```

Gameplay-affecting progression happens **inside the match** through standardized map pickups or other symmetric rules.

Examples:

- +1 bomb
- increased blast radius
- movement speed
- bomb kick
- remote detonator
- shield

Ranked maps should be designed or generated so starting positions have comparable opportunity. Randomness may exist, but competitive rules should avoid one spawn receiving a materially stronger opening solely by luck.

### Target arena rules

- 2-8 players depending on mode.
- Destructible and indestructible terrain.
- Bomb placement with readable fuse timing.
- Orthogonal blast propagation blocked by hard terrain.
- Short matches suitable for mobile.
- Server-authoritative or deterministically verifiable simulation for competitive play.
- Cosmetics never modify collision, timing, damage, visibility, or any other competitive statistic.

Candidate cosmetic bomb themes include Cherry Bomb, Carrot Charge, Trashcan TNT, Bee Bomb, Magnet Bomb, Burrow Bomb, Freeze Bomb, and Glitter Bomb. In ranked play, a skin that represents a standard bomb has the standard bomb's exact gameplay behavior.

## Missions

Missions are optional activities presented through NPC story/lore. They are the primary game-facing way to expose Proof of Play.

A mission may look like:

- Recon Patrol
- Secure the Supply Line
- Checkpoint Duty
- Communications Watch
- Warren Defense

Underneath the fiction, an eligible mission can correspond to useful network participation such as signing an unpredictable challenge, validating a checkpoint/state transition, participating in a committee, or other protocol work.

Completing missions can increase Proof-of-Play authority. Skipping missions must not stop ordinary idle progression or access to skill-based Warren Wars.

## Story progression

Use chapters/seasons/locations rather than endless anonymous levels. Early candidates:

1. Broken Burrow
2. Scrap Row
3. Carrot District
4. Downtown Warren
5. Officer Country
6. Billionaire Burrows
7. future absurd theaters

Seasonal story examples can include the Great Carrot Shortage, Operation Boomtown, a rival battalion, and a Warren Wars championship, but final story beats remain open for writing.

## MVP definition

The first game MVP should prove:

- player-created bunny profile
- one idle business with offline earnings
- one meaningful upgrade choice
- one NPC-driven story/tutorial path
- one deterministic local Warren Wars arena
- equal-start competitive rules
- one optional mission surface that can later bind to Proof of Play
