import { For, ParentProps } from "solid-js";
import FormInput from "./FormInput";

interface CheckboxElement {
  id: string | number;
  name: string;
}

type Orientation = "horizontal" | "vertical";

interface Props {
  id: string;
  title: string;
  orientation?: Orientation;
  class: string;
  elements: () => CheckboxElement[];
  value: (id: string | number) => string;
  onChange: (
    e: Event & {
      target: HTMLInputElement;
    },
    id: string | number,
  ) => void;
}

export default function CheckboxInput(props: Props & ParentProps) {
  const styles = () => ({
    "flex-col": props.orientation !== "horizontal",
    "flex-row": props.orientation === "horizontal",
    "gap-y-2": props.orientation !== "horizontal",
    "gap-x-8": props.orientation === "horizontal",
  });

  return (
    <FormInput id={props.id} title={props.title} class={props.class}>
      <div class="mt-2 flex" classList={styles()}>
        <For each={props.elements()}>
          {(element) => (
            <div class="flex flex-row items-center gap-3">
              <div class="group grid size-4 grid-cols-1">
                <input
                  id={element.id.toString()}
                  name={element.id.toString()}
                  type="checkbox"
                  class="col-start-1 row-start-1 appearance-none rounded border border-gray-300 bg-white checked:border-black checked:bg-black indeterminate:border-black indeterminate:bg-black focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-black disabled:border-gray-300 disabled:bg-gray-100 disabled:checked:bg-gray-100 forced-colors:appearance-auto"
                  value={props.elements().includes(element).toString()}
                  onChange={(e) => props.onChange(e, element.id.toString())}
                />
                <svg
                  class="pointer-events-none col-start-1 row-start-1 size-3.5 self-center justify-self-center stroke-white group-has-[:disabled]:stroke-gray-950/25"
                  viewBox="0 0 14 14"
                  fill="none"
                >
                  <path
                    class="opacity-0 group-has-[:checked]:opacity-100"
                    d="M3 8L6 11L11 3.5"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                  <path
                    class="opacity-0 group-has-[:indeterminate]:opacity-100"
                    d="M3 7H11"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
              </div>
              <div class="text-sm/6">
                <label
                  for={element.id.toString()}
                  class="text-sm/6 text-gray-600"
                >
                  {element.name}
                </label>
              </div>
            </div>
          )}
        </For>
      </div>
    </FormInput>
  );
}
