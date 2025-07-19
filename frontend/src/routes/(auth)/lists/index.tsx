import { A } from "@solidjs/router";
import { createResource, For, Show } from "solid-js";
import { getLists } from "~/lib/api";
import { List } from "~/lib/model";

export default function Index() {
	const [lists] = createResource(getLists);

	const domains = (list: List) => {
		if (list.domainIds.length === 0) {
			return "None";
		}

		return list.domainIds.length;
	};

	const schedules = (list: List) => {
		if (list.scheduleIds.length === 0) {
			return "None";
		}

		return list.scheduleIds.length;
	};

	return (
		<div class="mt-4 flow-root">
			<div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
				<div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
					<table class="min-w-full divide-y divide-gray-300">
						<thead>
							<tr>
								<th
									scope="col"
									class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-800 sm:pl-0"
								>
									Name
								</th>
								<th
									scope="col"
									class="px-3 py-3.5 text-left text-sm font-semibold text-gray-800"
									style=""
								>
									Domains
								</th>
								<th
									scope="col"
									class="px-3 py-3.5 text-left text-sm font-semibold text-gray-800"
								>
									Schedules
								</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-200">
							<Show when={lists()}>
								{(lists) => {
									console.log(lists());
									return (
										<For each={lists()}>
											{(list) => (
												<tr>
													<td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-800 sm:pl-0">
														<A
															href={`/lists/${list.id}`}
															class="cursor-pointer"
														>
															{list.name}
														</A>
													</td>
													<td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
														{domains(list)}
													</td>
													<td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
														{schedules(list)}
													</td>
												</tr>
											)}
										</For>
									);
								}}
							</Show>
						</tbody>
					</table>
				</div>
			</div>
		</div>
	);
}
