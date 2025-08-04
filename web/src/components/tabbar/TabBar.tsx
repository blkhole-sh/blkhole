import { A, useLocation, useNavigate } from "@solidjs/router";
import { For } from "solid-js";

export interface Tab {
	title: string;
	path: string;
}

interface Props {
	tabs: Tab[];
}

export default function TabBar(props: Props) {
	const location = useLocation();
	const navigate = useNavigate();

	const isActive = (path: string) => location.pathname.startsWith(path);

	const getCurrentTab = () => {
		return (
			props.tabs.find((tab) => location.pathname.startsWith(tab.path))?.path ??
			props.tabs[0].path
		);
	};

	const handleSelectChange = (event: Event) => {
		const target = event.target as HTMLSelectElement;
		navigate(target.value);
	};

	return (
		<>
			<div class="grid grid-cols-1 sm:hidden">
				<select
					aria-label="Select a tab"
					class="col-start-1 row-start-1 w-full appearance-none rounded-md bg-white py-2 pr-8 pl-3 text-base text-gray-900 outline-1 -outline-offset-1 outline-gray-300 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600"
					value={getCurrentTab()}
					onChange={handleSelectChange}
				>
					<For each={props.tabs}>
						{(tab) => <option value={tab.path}>{tab.title}</option>}
					</For>
				</select>
				<svg
					viewBox="0 0 16 16"
					fill="currentColor"
					data-slot="icon"
					aria-hidden="true"
					class="pointer-events-none col-start-1 row-start-1 mr-2 size-5 self-center justify-self-end fill-gray-500"
				>
					<path
						d="M4.22 6.22a.75.75 0 0 1 1.06 0L8 8.94l2.72-2.72a.75.75 0 1 1 1.06 1.06l-3.25 3.25a.75.75 0 0 1-1.06 0L4.22 7.28a.75.75 0 0 1 0-1.06Z"
						clip-rule="evenodd"
						fill-rule="evenodd"
					/>
				</svg>
			</div>
			<div class="hidden sm:block">
				<div class="border-b border-gray-200">
					<nav aria-label="Tabs" class="-mb-px flex space-x-8">
						<For each={props.tabs}>
							{(tab) => (
								<A
									href={tab.path}
									class="border-b-2 px-1 py-4 text-sm font-medium whitespace-nowrap"
									classList={{
										"border-black text-black": isActive(tab.path),
										"border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700":
											!isActive(tab.path),
									}}
								>
									{tab.title}
								</A>
							)}
						</For>
					</nav>
				</div>
			</div>
		</>
	);
}
