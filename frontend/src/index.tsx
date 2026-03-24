import {
  LocationProvider,
  Router,
  Route,
  lazy,
  hydrate,
  prerender as ssr,
} from "preact-iso";
import { AuthProvider } from "./contexts/AuthContext";
import "./i18n";

const Home = lazy(() => import("./pages/Home/index.jsx"));
const Register = lazy(() => import("./pages/Register/index.jsx"));
const Terms = lazy(() => import("./pages/Terms/index.jsx"));
const Privacy = lazy(() => import("./pages/Privacy/index.jsx"));
const Dashboard = lazy(() => import("./pages/Dashboard/index.jsx"));
const ChannelDetail = lazy(() => import("./pages/ChannelDetail/index.jsx"));
const Profile = lazy(() => import("./pages/Profile/index.jsx"));
const VideoPlayer = lazy(() => import("./pages/VideoPlayer/index.jsx"));
const Analytics = lazy(() => import("./pages/Analytics/index.jsx"));
const History = lazy(() => import("./pages/History/index.jsx"));
const Playlists = lazy(() => import("./pages/Playlists/index.jsx"));
const PlaylistDetail = lazy(() => import("./pages/PlaylistDetail/index.jsx"));
const Explore = lazy(() => import("./pages/Explore/index.jsx"));
const Search = lazy(() => import("./pages/Search/index.jsx"));
const NotFound = lazy(() => import("./pages/_404.jsx"));
import "./style.css";

function AppContent() {
  return (
    <Router>
      <Route path="/" component={Home} />
      <Route path="/dashboard" component={Dashboard} />
      <Route path="/channels" component={Explore} />
      <Route path="/channels/:channelId" component={ChannelDetail} />
      <Route path="/watch/:videoId" component={VideoPlayer} />
      <Route path="/analytics" component={Analytics} />
      <Route path="/history" component={History} />
      <Route path="/playlists" component={Playlists} />
      <Route path="/playlists/:playlistId" component={PlaylistDetail} />
      <Route path="/search" component={Search} />
      <Route path="/profile" component={Profile} />
      <Route path="/register" component={Register} />
      <Route path="/terms" component={Terms} />
      <Route path="/privacy" component={Privacy} />
      <Route default component={NotFound} />
    </Router>
  );
}

export function App() {
  return (
    <LocationProvider>
      <AuthProvider>
        <AppContent />
      </AuthProvider>
    </LocationProvider>
  );
}

if (typeof window !== "undefined") {
  hydrate(<App />, document.getElementById("app"));
}

export async function prerender(data) {
  return await ssr(<App {...data} />);
}
