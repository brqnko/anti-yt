import type { ColorMode } from "./hooks/useColorMode";

export const SITE_URL = "https://anti-yt.app";

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
  FEED: 50,
  PLAYLISTS: 25,
  HISTORY: 25,
  CHANNEL_VIDEOS: 25,
  CHANNEL_PLAYLISTS: 10,
  CHANNEL_PLAYLISTS_PAGE: 25,
  PLAYLIST_VIDEOS: 25,
  SEARCH: 50,
} as const;
