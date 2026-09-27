export function RegistrationProgress({ active }: { active: 1 | 2 | 3 }) {
  return (
    <ul className="steps steps-horizontal mb-5 w-full text-[.65rem]">
      <li className="step step-primary">Account</li>
      <li className={`step ${active >= 2 ? 'step-primary' : ''}`}>Verify</li>
      <li className={`step ${active >= 3 ? 'step-primary' : ''}`}>Welcome</li>
    </ul>
  )
}
