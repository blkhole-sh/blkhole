import LogoIcon from "./LogoIcon";

interface Props {
	message: string;
	subtitle: string;
}

export default function EmptyState(props: Props) {
	return (
		<div class="flex flex-1 items-center justify-center py-20">
			<div class="flex flex-col gap-4 max-w-md text-center">
				<LogoIcon class="w-8 h-8 mx-auto" />
				<p class="font-display text-xl">{props.message}</p>
				<p class="text-zinc-500 text-sm">{props.subtitle}</p>
			</div>
		</div>
	);
}
