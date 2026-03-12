import {
  LocationProvider,
  Router,
  Route,
  lazy,
  hydrate,
  prerender as ssr,
} from "preact-iso";
import type { ComponentChildren } from "preact";
import "./i18n";

import { Header } from "./components/Header.jsx";
const Home = lazy(() => import("./pages/Home/index.jsx"));
const Terms = lazy(() => import("./pages/Terms/index.jsx"));
const Privacy = lazy(() => import("./pages/Privacy/index.jsx"));
const NotFound = lazy(() => import("./pages/_404.jsx"));
import "./style.css";

/** Wraps non-standalone pages with the shared Header + main layout. */
function WithHeader(props: { children?: ComponentChildren }) {
  return (
    <>
      <Header />
      <main>{props.children}</main>
    </>
  );
}

function WrappedNotFound() {
  return (
    <WithHeader>
      <NotFound />
    </WithHeader>
  );
}

function AppContent() {
  return (
    <Router>
      <Route path="/" component={Home} />
      <Route path="/terms" component={Terms} />
      <Route path="/privacy" component={Privacy} />
      <Route default component={WrappedNotFound} />
    </Router>
  );
}

export function App() {
  return (
    <LocationProvider>
      <AppContent />
    </LocationProvider>
  );
}

if (typeof window !== "undefined") {
  hydrate(<App />, document.getElementById("app"));
}

export async function prerender(data) {
  return await ssr(<App {...data} />);
}
