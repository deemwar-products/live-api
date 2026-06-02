import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { LandingPage } from "@/pages/landing-page";
import { ThemeProvider } from "@/lib/use-theme";
import "@/styles/index.css";

function App() {
  return (
    <ThemeProvider>
      <LandingPage />
    </ThemeProvider>
  );
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
