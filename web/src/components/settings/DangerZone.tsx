interface Props {
	onDelete: () => void;
}

export default function DangerZone(props: Props) {
	return (
		<section class="flex flex-col gap-6">
			<h2 class="font-medium tracking-wider">DANGER ZONE</h2>
			<div class="flex flex-col gap-2">
				<button
					type="button"
					class="text-sm text-black font-medium underline tracking-wider hover:text-zinc-600 cursor-pointer self-start"
					onclick={props.onDelete}
				>
					DELETE ACCOUNT
				</button>
			</div>
		</section>
	);
}
