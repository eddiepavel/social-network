import type { InputHTMLAttributes, TextareaHTMLAttributes } from "react";

type BaseProps = {
  label: string;
  name: string;
};

type InputProps = BaseProps & InputHTMLAttributes<HTMLInputElement> & { as?: "input" };

type TextareaProps = BaseProps &
  TextareaHTMLAttributes<HTMLTextAreaElement> & { as: "textarea" };

export default function FormField(props: InputProps | TextareaProps) {
  if (props.as === "textarea") {
    const { label, name, as: _as, ...rest } = props;
    return (
      <label className="form-field">
        <span>{label}</span>
        <textarea name={name} {...rest} />
      </label>
    );
  }

  const { label, name, ...rest } = props;
  return (
    <label className="form-field">
      <span>{label}</span>
      <input name={name} {...rest} />
    </label>
  );
}
