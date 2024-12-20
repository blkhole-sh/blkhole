import { createResource, For, Show } from "solid-js";
import { getDevices } from "~/lib/api";
import { Device } from "~/lib/model";

export default function Index() {
  const [devices] = createResource(getDevices);

  const schedules = (device: Device) => {
    if (device.scheduleIds.length === 0) {
      return "None";
    }

    return device.scheduleIds.length;
  };

  return (
    <div class="py-6 px-8">
      <div class="sm:flex sm:items-center">
        <div class="sm:flex-auto">
          <h1 class="text-base font-semibold text-gray-800">Devices</h1>
          <p class="mt-2 text-sm text-gray-700">
            Please add all your devices here. Leo supports Apple (iOS, iPadOS,
            macOS), Android, Linux and Windows.
          </p>
        </div>
        <div class="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
          <a
            href="/devices/new"
            class="block rounded-md bg-black px-3 py-2 text-center text-sm font-semibold 
            text-white shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2"
          >
            Add
          </a>
        </div>
      </div>
      <div class="mt-8 flow-root">
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
                  >
                    Operating System
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
                <Show when={devices()}>
                  {(devices) => {
                    console.log(devices());

                    return (
                      <For each={devices()}>
                        {(device) => (
                          <tr>
                            <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-800 sm:pl-0">
                              <a
                                href={`/devices/${device.hash}`}
                                class="cursor-pointer"
                              >
                                {device.name}
                              </a>
                            </td>
                            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                              {device.os}
                            </td>
                            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                              {schedules(device)}
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
    </div>
  );
}
