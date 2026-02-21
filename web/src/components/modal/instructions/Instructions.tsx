import { ParentProps } from "solid-js";

function Title(props: ParentProps) {
	return <p class="text-sm text-zinc-700">{props.children}</p>;
}

function Instructions(props: ParentProps) {
	return (
		<div class="bg-zinc-50 p-4 flex flex-col gap-4 text-sm">
			<p class="mb-2 font-medium text-zinc-700 tracking-wider">
				SETUP INSTRUCTIONS
			</p>
			<ol class="list-decimal list-inside space-y-4 text-zinc-600">
				{props.children}
			</ol>
		</div>
	);
}

Instructions.Title = Title;
export default Instructions;
