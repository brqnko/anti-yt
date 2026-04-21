import {
  LocationProvider,
  Router,
  Route,
  lazy,
  hydrate,
  prerender as ssr,
} from "preact-iso";
import { SWRConfig } from "swr";
import { AuthProvider } from "./contexts/AuthContext";
import { ScreenTimeGate } from "./components/ScreenTimeGate";
import "./i18n";

const Register = lazy(() => import("./pages/Register/index.jsx"));
const Reactivation = lazy(() => import("./pages/Reactivation/index.jsx"));
const Terms = lazy(() => import("./pages/Terms/index.jsx"));
const Privacy = lazy(() => import("./pages/Privacy/index.jsx"));
const Dashboard = lazy(() => import("./pages/Dashboard/index.jsx"));
const ChannelDetail = lazy(() => import("./pages/ChannelDetail/index.jsx"));
const ChannelPlaylists = lazy(() => import("./pages/ChannelPlaylists/index.jsx"));
const Profile = lazy(() => import("./pages/Profile/index.jsx"));
const VideoPlayer = lazy(() => import("./pages/VideoPlayer/index.jsx"));
const Analytics = lazy(() => import("./pages/Analytics/index.jsx"));
const History = lazy(() => import("./pages/History/index.jsx"));
const Playlists = lazy(() => import("./pages/Playlists/index.jsx"));
const PlaylistDetail = lazy(() => import("./pages/PlaylistDetail/index.jsx"));
const Channels = lazy(() => import("./pages/Channels/index.jsx"));
const Explore = lazy(() => import("./pages/Explore/index.jsx"));
const Search = lazy(() => import("./pages/Search/index.jsx"));
const ScreenTimeSettings = lazy(() => import("./pages/ScreenTimeSettings/index.jsx"));
const NotFound = lazy(() => import("./pages/_404.jsx"));
import "./style.css";

function AppContent() {
  return (
    <Router>
      <Route path="/" component={Dashboard} />
      <Route path="/channels" component={Channels} />
      <Route path="/channels/explore" component={Explore} />
      <Route path="/channels/:channelId" component={ChannelDetail} />
      <Route path="/channels/:channelId/playlists" component={ChannelPlaylists} />
      <Route path="/watch/:videoId" component={VideoPlayer} />
      <Route path="/analytics" component={Analytics} />
      <Route path="/history" component={History} />
      <Route path="/playlists" component={Playlists} />
      <Route path="/playlists/:playlistId" component={PlaylistDetail} />
      <Route path="/search" component={Search} />
      <Route path="/profile" component={Profile} />
      <Route path="/screen-time-settings" component={ScreenTimeSettings} />
      <Route path="/register" component={Register} />
      <Route path="/reactivation" component={Reactivation} />
      <Route path="/terms" component={Terms} />
      <Route path="/privacy" component={Privacy} />
      <Route default component={NotFound} />
    </Router>
  );
}

export function App() {
  return (
    <LocationProvider>
      <SWRConfig
        value={{
          revalidateOnFocus: true,
          revalidateOnReconnect: true,
          dedupingInterval: 30_000,
          focusThrottleInterval: 10_000,
          keepPreviousData: true,
          errorRetryCount: 1,
        }}
      >
        <AuthProvider>
          <ScreenTimeGate>
            <AppContent />
          </ScreenTimeGate>
        </AuthProvider>
      </SWRConfig>
    </LocationProvider>
  );
}

if (typeof window !== "undefined") {
  const el = document.getElementById("app");
  if (el) hydrate(<App />, el);
}

export async function prerender(data: Record<string, unknown>) {
  return await ssr(<App {...data} />);
}
