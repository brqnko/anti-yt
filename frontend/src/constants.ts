import type { ColorMode } from "./hooks/useColorMode";

export const modeIcons: Record<ColorMode, string> = {
  light: "light_mode",
  dark: "dark_mode",
  system: "computer",
};

export const modeOrder: ColorMode[] = ["light", "dark", "system"];

export const languages = [
  { code: "ja", label: "日本語" },
  { code: "en", label: "English" },
];

export const REPORT_FORM_URL = "https://docs.google.com/forms/d/e/1FAIpQLSfbAk6J4jSW83eKe1CeGCCqqHI517STCPh3hE2zRbWedbdcCQ/viewform?usp=dialog";

export const PAGE_SIZES = {
  FEED: 12,
  PLAYLISTS: 12,
  HISTORY: 20,
  CHANNEL_VIDEOS: 30,
  PLAYLIST_VIDEOS: 20,
  SEARCH: 50,
} as const;
