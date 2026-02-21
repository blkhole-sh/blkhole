import { JSX, ParentProps, Show } from "solid-js";
import Divider from "../ui/Divider";
import ChevronDown from "../icons/ChevronDown";
import { cx } from "~/lib/utils";

interface Props extends ParentProps {
	label: string;
	value: string;
	onChange?: JSX.EventHandlerUnion<HTMLSelectElement, Event>;
	placeholder?: string;
	hint?: string;
	class?: string;
}

export default function SelectInput(props: Props) {
	return (
		<div class={cx("flex flex-col gap-1", props.class)}>
			<label
				for={props.label}
				class="font-medium text-zinc-700 text-sm tracking-wider"
			>
				{props.label}
			</label>
			<div class="relative w-full">
				<select
					id={props.label}
					value={props.value}
					onChange={props.onChange}
					class="w-full py-2 pr-8 text-sm tracking-wider outline-none bg-transparent appearance-none cursor-pointer"
					classList={{ "text-zinc-400": props.value === "" }}
				>
					<Show when={props.placeholder}>
						<option value="" disabled hidden>
							{props.placeholder}
						</option>
					</Show>
					{props.children}
				</select>
				<ChevronDown class="size-4 absolute right-1 top-1/2 -translate-y-1/2 text-zinc-400 pointer-events-none" />
			</div>
			<Divider />
			<Show when={props.hint}>
				<p class="text-xs text-zinc-400">{props.hint}</p>
			</Show>
		</div>
	);
}
