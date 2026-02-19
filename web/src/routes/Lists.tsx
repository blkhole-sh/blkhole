import { createResource, For } from "solid-js";
import ListTile from "~/components/ListTile";
import PageShell from "~/components/PageShell";
import { getLists } from "~/lib/api";

export default function Lists() {
	const [lists] = createResource(getLists);

	return (
		<PageShell
			title="Blocklists"
			description="Curated domain lists used to filter unwanted content across your network."
			cta="ADD BLOCKLIST"
		>
			<div class="divide-y divide-zinc-200">
				<For each={lists()}>{(list) => <ListTile list={list} />}</For>
			</div>
		</PageShell>
	);
}
