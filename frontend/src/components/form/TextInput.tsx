import { ParentProps } from "solid-js";
import FormInput from "./FormInput";

interface Props {
  id: string;
  title: string;
  placeholder: string;
  class: string;
  value: () => string;
  onChange: (
    e: Event & {
      target: HTMLInputElement;
    },
  ) => void;
}

export default function TextInput(props: Props & ParentProps) {
  return (
    <FormInput id={props.id} title={props.title} class={props.class}>
      <div class="flex items-center rounded-md bg-white pl-3 outline 
                  -outline-offset-1 outline-gray-300 focus-within:outline 
                  focus-within:outline-black">
        <input
          type="text"
          name={props.id}
          id={props.id}
          class="block min-w-0 grow py-1.5 pl-1 pr-3 text-base text-gray-900 placeholder:text-gray-400 focus:outline  sm:text-sm/6"
          placeholder={props.placeholder}
          value={props.value()}
          onChange={props.onChange}
        />
      </div>
    </FormInput>
  );
}
