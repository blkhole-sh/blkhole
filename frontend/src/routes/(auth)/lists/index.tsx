import { createResource, For, Show } from "solid-js";
import ListTile from "~/components/list/ListTile";
import { getLists } from "~/lib/api";

export default function Index() {
	const [lists] = createResource(getLists);

	return (
		<div class="mt-2 flow-root">
			<div class="inline-block min-w-full align-middle">
				<ul role="list" class="divide-y divide-gray-100">
					<For each={lists()}>{(list) => <ListTile list={list} />}</For>
				</ul>
			</div>
		</div>
	);
}
