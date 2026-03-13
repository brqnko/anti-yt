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
const Profile = lazy(() => import("./pages/Profile/index.jsx"));
const NotFound = lazy(() => import("./pages/_404.jsx"));
import "./style.css";

function AppContent() {
  return (
    <Router>
      <Route path="/" component={Home} />
      <Route path="/dashboard" component={Dashboard} />
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
