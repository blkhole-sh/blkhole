import { ParentProps } from "solid-js";

interface Props {
  id: string;
  title: string;
  class: string;
}

export default function FormInput(props: Props & ParentProps) {
  return (
    <div class={props.class}>
      <label for={props.id} class="block text-sm/6 font-medium text-gray-900">
        {props.title}
      </label>
      <div class="mt-2">{props.children}</div>
    </div>
  );
}
