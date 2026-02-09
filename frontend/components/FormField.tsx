import type { InputHTMLAttributes, TextareaHTMLAttributes } from "react";

type BaseProps = {
  label: string;
  name: string;
  error?: string;
};

type InputProps = BaseProps & InputHTMLAttributes<HTMLInputElement> & { as?: "input" };

type TextareaProps = BaseProps &
  TextareaHTMLAttributes<HTMLTextAreaElement> & { as: "textarea" };

export default function FormField(props: InputProps | TextareaProps) {
  if (props.as === "textarea") {
    const { label, name, error, as: _as, ...rest } = props;
    return (
      <label className="form-field">
        <span>{label}</span>
        <textarea
          name={name}
          {...rest}
          style={{
            borderColor: error ? "#b42318" : undefined,
            ...rest.style,
          }}
        />
        {error && <span style={{ color: "#b42318", fontSize: "0.85rem" }}>{error}</span>}
      </label>
    );
  }

  const { label, name, error, ...rest } = props;
  return (
    <label className="form-field">
      <span>{label}</span>
      <input
        name={name}
        {...rest}
        style={{
          borderColor: error ? "#b42318" : undefined,
          ...rest.style,
        }}
      />
      {error && <span style={{ color: "#b42318", fontSize: "0.85rem" }}>{error}</span>}
    </label>
  );
}
