import {
	children,
	createEffect,
	createSignal,
	JSX,
	ParentProps,
	Show,
} from "solid-js";
import ActionButton from "../ui/ActionButton";
import XMark from "../icons/XMark";
import ButtonGhost from "../ui/ButtonGhost";
import ButtonSolid from "../ui/ButtonSolid";

function Step(props: ParentProps) {
	return <>{props.children}</>;
}

interface Props extends ParentProps {
	title: string;
	open: boolean;
	onClose: () => void;
	onConfirm: () => void;
	onBeforeNext?: (step: number) => boolean | Promise<boolean>;
}

function MultiStepModal(props: Props) {
	let ref!: HTMLDialogElement;
	const [step, setStep] = createSignal(0);
	const resolved = children(() => props.children);

	createEffect(() => {
		if (props.open && !ref.open) {
			ref.showModal();
			// Focus the dialog itself to prevent auto-focus on close button
			setTimeout(() => ref.focus(), 0);
		} else if (!props.open && ref.open) {
			ref.close();
		}
	});

	createEffect(() => {
		if (!props.open) setStep(0);
	});

	const steps = () => resolved.toArray() as JSX.Element[];
	const isFirst = () => step() === 0;
	const isLast = () => step() === steps().length - 1;

	const handleNext = async () => {
		const result = await props.onBeforeNext?.(step());
		if (result === false) return;
		setStep((s) => s + 1);
	};

	return (
		<dialog
			ref={ref}
			onClose={props.onClose}
			class="w-full max-w-lg max-h-[90vh] open:flex open:flex-col divide-y divide-zinc-100 open:fixed open:top-1/2 open:left-1/2 open:-translate-x-1/2 open:-translate-y-1/2 backdrop:bg-black/30 backdrop:backdrop-blur-sm outline-none"
		>
			<div class="p-8 flex flex-row justify-between items-start flex-shrink-0">
				<div class="flex flex-col gap-1">
					<p class="text-xs text-zinc-400 tracking-wider">
						STEP {step() + 1} OF {steps().length}
					</p>
					<h2 class="font-display text-2xl">{props.title}</h2>
				</div>
				<ActionButton onclick={props.onClose} tabindex={-1}>
					<XMark />
				</ActionButton>
			</div>
			<div class="p-8 overflow-y-auto">{steps()[step()]}</div>
			<div class="px-8 py-6 flex flex-row justify-end items-center gap-6 w-full flex-shrink-0">
				<Show when={isFirst()}>
					<ButtonGhost onclick={props.onClose}>CANCEL</ButtonGhost>
				</Show>
				<Show when={!isFirst()}>
					<ButtonGhost onclick={() => setStep((s) => s - 1)}>BACK</ButtonGhost>
				</Show>
				<Show when={!isLast()}>
					<ButtonSolid onclick={handleNext}>NEXT</ButtonSolid>
				</Show>
				<Show when={isLast()}>
					<ButtonSolid onclick={props.onConfirm}>CONFIRM</ButtonSolid>
				</Show>
			</div>
		</dialog>
	);
}

MultiStepModal.Step = Step;

export default MultiStepModal;
