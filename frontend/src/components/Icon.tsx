import Add from "~icons/material-symbols/add";
import AddCircle from "~icons/material-symbols/add-circle-outline";
import Analytics from "~icons/material-symbols/analytics-outline";
import ArrowBack from "~icons/material-symbols/arrow-back";
import ArrowForward from "~icons/material-symbols/arrow-forward";
import CalendarClock from "~icons/material-symbols/calendar-clock-outline";
import Check from "~icons/material-symbols/check";
import CheckCircle from "~icons/material-symbols/check-circle-outline";
import Close from "~icons/material-symbols/close";
import Computer from "~icons/material-symbols/computer-outline";
import ContentPaste from "~icons/material-symbols/content-paste";
import DarkMode from "~icons/material-symbols/dark-mode-outline";
import Delete from "~icons/material-symbols/delete-outline";
import Devices from "~icons/material-symbols/devices-outline";
import Edit from "~icons/material-symbols/edit-outline";
import EditNote from "~icons/material-symbols/edit-note-outline";
import Error from "~icons/material-symbols/error-outline";
import ExpandMore from "~icons/material-symbols/expand-more";
import Flag from "~icons/material-symbols/flag-outline";
import FormatBold from "~icons/material-symbols/format-bold";
import Fullscreen from "~icons/material-symbols/fullscreen";
import FullscreenExit from "~icons/material-symbols/fullscreen-exit";
import FormatItalic from "~icons/material-symbols/format-italic";
import FormatListBulleted from "~icons/material-symbols/format-list-bulleted";
import Gavel from "~icons/material-symbols/gavel";
import Grass from "~icons/material-symbols/grass";
import GridView from "~icons/material-symbols/grid-view-outline";
import History from "~icons/material-symbols/history";
import Home from "~icons/material-symbols/home-outline";
import Language from "~icons/material-symbols/language";
import LightMode from "~icons/material-symbols/light-mode-outline";
import Logout from "~icons/material-symbols/logout";
import Memory from "~icons/material-symbols/memory-outline";
import Menu from "~icons/material-symbols/menu";
import MusicNote from "~icons/material-symbols/music-note";
import Pause from "~icons/material-symbols/pause-outline";
import Person from "~icons/material-symbols/person-outline";
import PlayArrow from "~icons/material-symbols/play-arrow-outline";
import PlaylistPlay from "~icons/material-symbols/playlist-play";
import PlaylistRemove from "~icons/material-symbols/playlist-remove";
import ProgressActivity from "~icons/material-symbols/progress-activity";
import Recommend from "~icons/material-symbols/recommend-outline";
import Refresh from "~icons/material-symbols/refresh";
import Schedule from "~icons/material-symbols/schedule-outline";
import School from "~icons/material-symbols/school-outline";
import Search from "~icons/material-symbols/search";
import SearchOff from "~icons/material-symbols/search-off";
import Subscriptions from "~icons/material-symbols/subscriptions-outline";
import Timer from "~icons/material-symbols/timer-outline";
import Translate from "~icons/material-symbols/translate";
import Tune from "~icons/material-symbols/tune";
import TrendingUp from "~icons/material-symbols/trending-up";
import VideocamOff from "~icons/material-symbols/videocam-off-outline";
import VolumeDown from "~icons/material-symbols/volume-down-outline";
import VolumeOff from "~icons/material-symbols/volume-off-outline";
import VolumeUp from "~icons/material-symbols/volume-up-outline";
import Warning from "~icons/material-symbols/warning-outline";

const iconMap: Record<string, string> = {
  add: Add,
  add_circle: AddCircle,
  analytics: Analytics,
  arrow_back: ArrowBack,
  arrow_forward: ArrowForward,
  calendar_clock: CalendarClock,
  check: Check,
  check_circle: CheckCircle,
  close: Close,
  computer: Computer,
  content_paste: ContentPaste,
  dark_mode: DarkMode,
  delete: Delete,
  devices: Devices,
  edit: Edit,
  edit_note: EditNote,
  error: Error,
  error_outline: Error,
  expand_more: ExpandMore,
  flag: Flag,
  format_bold: FormatBold,
  format_italic: FormatItalic,
  format_list_bulleted: FormatListBulleted,
  fullscreen: Fullscreen,
  fullscreen_exit: FullscreenExit,
  gavel: Gavel,
  grass: Grass,
  grid_view: GridView,
  history: History,
  home: Home,
  language: Language,
  light_mode: LightMode,
  logout: Logout,
  memory: Memory,
  menu: Menu,
  music_note: MusicNote,
  pause: Pause,
  person: Person,
  play_arrow: PlayArrow,
  playlist_play: PlaylistPlay,
  playlist_remove: PlaylistRemove,
  progress_activity: ProgressActivity,
  recommend: Recommend,
  refresh: Refresh,
  schedule: Schedule,
  school: School,
  search: Search,
  search_off: SearchOff,
  subscriptions: Subscriptions,
  timer: Timer,
  translate: Translate,
  tune: Tune,
  trending_up: TrendingUp,
  videocam_off: VideocamOff,
  volume_down: VolumeDown,
  volume_off: VolumeOff,
  volume_up: VolumeUp,
  warning: Warning,
};

interface IconProps {
  name: string;
  class?: string;
}

export function Icon({ name, class: className }: IconProps) {
  const svg = iconMap[name];
  if (!svg) return null;
  return (
    <span
      class={className}
      style={{ display: "inline-flex", verticalAlign: "middle" }}
      dangerouslySetInnerHTML={{ __html: svg }}
    />
  );
}
