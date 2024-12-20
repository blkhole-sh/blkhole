import SideBarItem from "./SideBarItem";

export default function SideBar() {
	return (
		<div class="flex grow flex-col gap-y-4 p-4 overflow-y-auto border-r border-gray-200 bg-white h-screen">
			<nav class="flex flex-1 flex-col">
				<ul role="list" class="flex flex-1 flex-col gap-y-6">
					<li>
						<ul role="list" class="space-y-1">
							<SideBarItem icon="🗓️" title="Schedules" url="/schedules" />
							<SideBarItem icon="📝" title="Lists" url="/lists" />
							<SideBarItem icon="📱" title="Devices" url="/devices" />
						</ul>
					</li>
					<li class="mt-auto flex flex-col">
						<div
							class="cursor-pointer flex flex-1 p-2 text-sm text-gray-600 font-medium hover:bg-gray-50"
							onClick={() => {}}
						>
							Logout
						</div>
					</li>
				</ul>
			</nav>
		</div>
	);
}
