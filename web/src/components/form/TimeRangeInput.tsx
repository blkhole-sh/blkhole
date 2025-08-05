import { ParentProps } from "solid-js";
import FormInput from "./FormInput";

interface Props {
	id: string;
	title: string;
	placeholderStart: string;
	placeholderEnd: string;
	class: string;
	valueStart: () => string;
	valueEnd: () => string;
	onChangeStart: (
		e: Event & {
			target: HTMLInputElement;
		},
	) => void;
	onChangeEnd: (
		e: Event & {
			target: HTMLInputElement;
		},
	) => void;
}

export default function TimeRangeInput(props: Props & ParentProps) {
	return (
		<FormInput id={props.id} title={props.title} class={props.class}>
			<div class="flex flex-row gap-x-8 items-center text-sm/6 text-gray-600">
				<div class="flex flex-row gap-x-2 items-center">
					Start
					<input
						type="time"
						name={`${props.id}-start`}
						id={`${props.id}-start`}
						class="block rounded-md border border-gray-300 bg-white px-3 py-1.5 text-base text-gray-900 placeholder:text-gray-400 sm:text-sm/6"
						placeholder={props.placeholderStart}
						value={props.valueStart()}
						onChange={props.onChangeStart}
					/>
				</div>
				<div class="flex flex-row gap-x-2 items-center">
					End
					<input
						type="time"
						name={`${props.id}-end`}
						id={`${props.id}-end`}
						class="block rounded-md border border-gray-300 bg-white px-3 py-1.5 text-base text-gray-900 placeholder:text-gray-400 sm:text-sm/6"
						placeholder={props.placeholderEnd}
						value={props.valueEnd()}
						onChange={props.onChangeEnd}
					/>
				</div>
			</div>
		</FormInput>
	);
}
