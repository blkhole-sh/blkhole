import type { JSX, ParentProps } from "solid-js";
import ButtonSolid from "~/components/ui/ButtonSolid";
import Logo from "~/components/ui/Logo";

interface Props extends ParentProps {
	onSubmit: (e: Event) => void;
	submitLabel: string;
	footer?: JSX.Element;
}

export default function AuthForm(props: Props) {
	return (
		<main class="min-h-screen flex items-center justify-center">
			<div class="w-sm">
				<Logo class="h-16 block mx-auto" />
				<form onSubmit={props.onSubmit} class="mt-6">
					<div class="flex flex-col gap-8 py-16 tracking-wider">
						{props.children}
					</div>
					<ButtonSolid
						type="submit"
						class="w-full my-1 py-4 text-base tracking-wider"
					>
						{props.submitLabel}
					</ButtonSolid>
					{props.footer && (
						<div class="h-16 flex flex-col gap-2 mt-4 text-sm text-zinc-500 items-center">
							{props.footer}
						</div>
					)}
				</form>
			</div>
		</main>
	);
}
