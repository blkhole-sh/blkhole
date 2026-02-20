import { ParentProps, createEffect } from "solid-js";
import ActionButton from "../ui/ActionButton";
import XMark from "../icons/XMark";
import ButtonGhost from "../ui/ButtonGhost";
import ButtonSolid from "../ui/ButtonSolid";

interface Props extends ParentProps {
	title: string;
	open: boolean;
	onClose: () => void;
	onConfirm: () => void;
	confirmLabel?: string;
}

export default function Modal(props: Props) {
	let ref!: HTMLDialogElement;

	createEffect(() => {
		if (props.open && !ref.open) {
			ref.showModal();
			// Focus the dialog itself to prevent auto-focus on close button
			setTimeout(() => ref.focus(), 0);
		} else if (!props.open && ref.open) {
			ref.close();
		}
	});

	return (
		<dialog
			ref={ref}
			onClose={props.onClose}
			class="w-full max-w-lg max-h-[90vh] open:flex open:flex-col divide-y divide-zinc-100 open:fixed open:top-1/2 open:left-1/2 open:-translate-x-1/2 open:-translate-y-1/2 backdrop:bg-black/30 backdrop:backdrop-blur-sm outline-none"
		>
			<div class="p-8 flex flex-row justify-between items-start flex-shrink-0">
				<h2 class="font-display text-2xl">{props.title}</h2>
				<ActionButton onclick={props.onClose} tabindex={-1}>
					<XMark />
				</ActionButton>
			</div>
			<div class="p-8 overflow-y-auto">{props.children}</div>
			<div class="px-8 py-6 flex flex-row justify-end items-center gap-6 w-full flex-shrink-0">
				<ButtonGhost onclick={props.onClose}>CANCEL</ButtonGhost>
				<ButtonSolid onclick={props.onConfirm}>
					{props.confirmLabel ?? "CONFIRM"}
				</ButtonSolid>
			</div>
		</dialog>
	);
}
