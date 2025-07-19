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
					<div class="flex items-center rounded-md bg-white pl-3 outline -outline-offset-1 outline-gray-300 focus-within:outline focus-within:-outline-offset-2 focus-within:outline-black">
						<input
							type="time"
							name={props.id}
							id={props.id}
							class="block min-w-0 grow pl-1 pr-3 text-base text-gray-900 placeholder:text-gray-400 focus:outline sm:text-sm/6"
							placeholder={props.placeholderStart}
							value={props.valueStart()}
							onChange={props.onChangeEnd}
						/>
					</div>
				</div>
				<div class="flex flex-row gap-x-2 items-center">
					End
					<div class="flex items-center rounded-md bg-white pl-3 outline -outline-offset-1 outline-gray-300 focus-within:outline focus-within:-outline-offset-2 focus-within:outline-black">
						<input
							type="time"
							name={props.id}
							id={props.id}
							class="block min-w-0 grow pl-1 pr-3 text-base text-gray-900 placeholder:text-gray-400 focus:outline sm:text-sm/6"
							placeholder={props.placeholderEnd}
							value={props.valueEnd()}
							onChange={props.onChangeEnd}
						/>
					</div>
				</div>
			</div>
		</FormInput>
	);
}
