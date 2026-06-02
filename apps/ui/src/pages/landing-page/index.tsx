import { Header } from "@/pages/landing-page/sections/header";
import { Hero } from "@/pages/landing-page/sections/hero";
import { Ticker } from "@/pages/landing-page/sections/ticker";
import { ProblemSection } from "@/pages/landing-page/sections/problem-section";
import { ProductSection } from "@/pages/landing-page/sections/product-section";
import { CapabilitiesSection } from "@/pages/landing-page/sections/capabilities-section";
import { LiveDemoSection } from "@/pages/landing-page/sections/live-demo-section";
import { UseCasesSection } from "@/pages/landing-page/sections/use-cases-section";
import { OutcomesSection } from "@/pages/landing-page/sections/outcomes-section";
import { PricingSection } from "@/pages/landing-page/sections/pricing-section";
import { CtaSection } from "@/pages/landing-page/sections/cta-section";
import { Footer } from "@/pages/landing-page/sections/footer";

export function LandingPage() {
  return (
    <div className="min-h-full">
      <Header />
      <main>
        <Hero />
        <Ticker />
        <ProblemSection />
        <ProductSection />
        <CapabilitiesSection />
        <LiveDemoSection />
        <UseCasesSection />
        <OutcomesSection />
        <PricingSection />
        <CtaSection />
      </main>
      <Footer />
    </div>
  );
}
