import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { LandingPage } from "@/pages/landing-page";
import { AuthPage } from "@/pages/auth-page";
import { ConsolePage } from "@/pages/console";
import { OnboardingPage } from "@/pages/console/onboarding";
import { LiveSessionPage } from "@/pages/live-session";
import { ThemeProvider } from "@/lib/use-theme";
import "@/styles/index.css";

function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<LandingPage />} />
          <Route path="/auth" element={<AuthPage />} />
          <Route path="/console" element={<ConsolePage />} />
          <Route path="/console/onboarding" element={<OnboardingPage />} />
 <Route path="/console/live" element={<LiveSessionPage />} />
        </Routes>
      </BrowserRouter>
    </ThemeProvider>
  );
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
