import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft, ArrowRight } from "lucide-react";
import { motion } from "framer-motion";
import { OnboardingShell } from "@/components/layout/onboarding-shell";
import { Button } from "@/components/ui/button";
import { StepKnowledge } from "@/pages/console/onboarding/parts/step-knowledge";
import { StepSystemPrompt } from "@/pages/console/onboarding/parts/step-system-prompt";
import { StepMcpTools } from "@/pages/console/onboarding/parts/step-mcp-tools";
import { StepInviteTeam } from "@/pages/console/onboarding/parts/step-invite-team";
import { StepSummary } from "@/pages/console/onboarding/parts/step-summary";
import { ONBOARDING_LABELS } from "@/labels";
import { CONSOLE } from "@/lib/paths";
import {
  getOnboardingState,
  markOnboardingComplete,
  type DocumentRow,
  type OnboardingState,
} from "@/mocks/business/onboarding";
import { ONBOARDING_PAGE_LABELS } from "@/pages/console/onboarding/labels";

const STEPS = [
  { id: 1, label: ONBOARDING_LABELS.stepper.step1 },
  { id: 2, label: ONBOARDING_LABELS.stepper.step2 },
  { id: 3, label: ONBOARDING_LABELS.stepper.step3 },
  { id: 4, label: ONBOARDING_LABELS.stepper.step4 },
];

const ease = [0.22, 1, 0.36, 1] as const;

export function OnboardingPage() {
  const navigate = useNavigate();
  const [state, setState] = useState<OnboardingState>(() => getOnboardingState());
  const [step, setStep] = useState<number>(1);

  useEffect(() => {
    document.title = ONBOARDING_PAGE_LABELS.meta.title;
  }, []);

  const onChange = (patch: Partial<OnboardingState>) => {
    setState((prev) => ({ ...prev, ...patch }));
  };

  const onAddDocument = (doc: DocumentRow) => {
    setState((prev) => ({ ...prev, documents: [...prev.documents, doc] }));
  };

  const goToStep = (id: number) => {
    if (id < 1 || id > STEPS.length) return;
    if (id > step && !state.completedSteps.includes(step)) {
      // Forward jumps only allowed past completed steps
      if (id !== step + 1) return;
    }
    setStep(id);
  };

  const handleNext = () => {
    if (step === STEPS.length) return;
    setState((prev) => ({
      ...prev,
      completedSteps: prev.completedSteps.includes(step) ? prev.completedSteps : [...prev.completedSteps, step],
    }));
    setStep((s) => s + 1);
  };

  const handleBack = () => {
    if (step === 1) return;
    setStep((s) => s - 1);
  };

  const handleSkip = () => {
    handleNext();
  };

  const handleFinish = () => {
    markOnboardingComplete();
    setState((prev) => ({
      ...prev,
      completedSteps: STEPS.map((s) => s.id),
    }));
    setStep(STEPS.length + 1);
  };

  const handleGoToDashboard = () => {
    markOnboardingComplete();
    navigate(CONSOLE.home);
  };

  const handleCustomize = () => {
    markOnboardingComplete();
    navigate(CONSOLE.settings);
  };

  const handleSaveAndExit = () => {
    navigate(CONSOLE.home);
  };

  const onSummary = step > STEPS.length;
  const isLast = step === STEPS.length;

  return (
    <OnboardingShell
      steps={STEPS}
      current={Math.min(step, STEPS.length)}
      completed={state.completedSteps}
      onStepClick={goToStep}
      onSaveAndExit={onSummary ? undefined : handleSaveAndExit}
    >
      <motion.div
        key={onSummary ? "summary" : step}
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease }}
      >
        {onSummary ? (
          <StepSummary onGoToDashboard={handleGoToDashboard} onCustomize={handleCustomize} />
        ) : (
          <>
            {step === 1 && <StepKnowledge state={state} onAddDocument={onAddDocument} />}
            {step === 2 && <StepSystemPrompt state={state} onChange={onChange} />}
            {step === 3 && <StepMcpTools state={state} onChange={onChange} />}
            {step === 4 && <StepInviteTeam state={state} onChange={onChange} />}

            <div className="mt-8 flex items-center justify-between gap-3 border-t border-border/60 pt-6">
              <Button
                variant="ghost"
                onClick={handleBack}
                disabled={step === 1}
                leading={<ArrowLeft className="size-4" />}
              >
                {ONBOARDING_LABELS.actions.back}
              </Button>
              <div className="flex items-center gap-2">
                {!isLast && (
                  <Button variant="secondary" onClick={handleSkip}>
                    {ONBOARDING_LABELS.actions.skip}
                  </Button>
                )}
                {isLast ? (
                  <Button onClick={handleFinish} trailing={<ArrowRight className="size-4" />}>
                    {ONBOARDING_LABELS.actions.finish}
                  </Button>
                ) : (
                  <Button onClick={handleNext} trailing={<ArrowRight className="size-4" />}>
                    {ONBOARDING_LABELS.actions.next}
                  </Button>
                )}
              </div>
            </div>
          </>
        )}
      </motion.div>
    </OnboardingShell>
  );
}
