export type SquadMember = {
  callsign: string;
  rank: string;
  role: string;
  quote: string;
  perk: string;
  icon: string;
};

export const squad: SquadMember[] = [
  {
    callsign: "Hard-as-Nails",
    rank: "First Sergeant",
    role: "Warren discipline",
    quote: "You call that a detonation? I've seen boiled carrots hit harder.",
    perk: "+15% workshop output",
    icon: "🎖️",
  },
  {
    callsign: "Da Champ",
    rank: "Specialist",
    role: "Arena celebrity",
    quote: "When Da Champ enters the arena, the money starts saluting.",
    perk: "+10% arena winnings",
    icon: "🏆",
  },
  {
    callsign: "Stuffy",
    rank: "Private",
    role: "Scrounger",
    quote: "Sir, respectfully, I think that bomb is hissing at me.",
    perk: "+20% scrap finds",
    icon: "🪖",
  },
  {
    callsign: "Boomboom",
    rank: "Corporal",
    role: "Demolitions",
    quote: "If it doesn't explode, I consider that unfinished.",
    perk: "Special bomb crafting",
    icon: "💣",
  },
  {
    callsign: "Cashmere",
    rank: "Captain",
    role: "Logistics & finance",
    quote: "Every operation needs an exit strategy and a margin.",
    perk: "+12% passive cash",
    icon: "💵",
  },
  {
    callsign: "Flopsy",
    rank: "Doc",
    role: "Combat medic",
    quote: "I can patch the fur. I can't patch bad decisions.",
    perk: "Faster squad recovery",
    icon: "🩹",
  },
];
