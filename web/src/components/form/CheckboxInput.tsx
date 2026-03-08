import { JSX } from "solid-js";

interface Props {
	label: string;
	description?: string;
	icon?: JSX.Element;
	checked: boolean;
	onChange: () => void;
}

export default function CheckboxInput(props: Props) {
	return (
		<label class="flex flex-row items-center justify-between gap-4 py-3 cursor-pointer">
			<div class="flex flex-row items-center gap-3">
				{props.icon && <div class="text-zinc-400 shrink-0">{props.icon}</div>}
				<div class="flex flex-col">
					<span class="font-medium text-black text-zinc-700 text-sm tracking-wider">
						{props.label}
					</span>
					{props.description && (
						<span class="text-xs text-zinc-400 max-w-md ">
							{props.description}
						</span>
					)}
				</div>
			</div>
			<input
				type="checkbox"
				checked={props.checked}
				onChange={props.onChange}
				class="cursor-pointer"
			/>
		</label>
	);
}
