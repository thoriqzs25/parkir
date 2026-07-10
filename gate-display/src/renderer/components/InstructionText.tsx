interface Props {
  text: string
}

export default function InstructionText({ text }: Props) {
  if (!text) return null
  return <div className="instruction">{text}</div>
}
