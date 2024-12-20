import { createResource, For, Show } from "solid-js";
import { useAuth } from "~/context/AuthContext";
import { getSchedules } from "~/lib/api";
import { Schedule } from "~/lib/model";

const activeDays = (schedule: Schedule): string => {
  const daysMap = {
    monday: "Mon",
    tuesday: "Tue",
    wednesday: "Wed",
    thursday: "Thu",
    friday: "Fri",
    saturday: "Sat",
    sunday: "Sun",
  };

  const days = Object.entries(daysMap)
    .filter(([dayKey]) => schedule[dayKey as keyof Schedule])
    .map(([_, abbreviation]) => abbreviation);

  return days.join(", ");
};

const listsAndDomains = (schedule: Schedule) => {
  const lists = schedule.listIds.length;
  const domains = schedule.domainIds.length;

  if (lists === 0 && domains === 0) {
    return "None"; // Optional: handle case when both are empty
  }

  if (lists === 0) {
    return `${domains} domains`;
  }

  if (domains === 0) {
    return `${lists} lists`;
  }

  return `${lists} lists, ${domains} domains`;
};

export default function Index() {
  const [schedules] = createResource(getSchedules);

  return (
    <div class="py-6 px-8">
      <div class="sm:flex sm:items-center">
        <div class="sm:flex-auto">
          <h1 class="text-base font-semibold text-gray-800">Schedules</h1>
          <p class="mt-2 text-sm text-gray-700">
            Here you find all your schedules.
          </p>
        </div>
        <div class="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
          <a
            href="/schedules/new"
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
                    class="px-3 py-3.5 text-left text-sm font-semibold text-gray-800" style=""
                  >
                    Start
                  </th>
                  <th
                    scope="col"
                    class="px-3 py-3.5 text-left text-sm font-semibold text-gray-800"
                  >
                    End
                  </th>
                  <th
                    scope="col"
                    class="px-3 py-3.5 text-left text-sm font-semibold text-gray-800"
                  >
                    Days
                  </th>
                  <th
                    scope="col"
                    class="px-3 py-3.5 text-left text-sm font-semibold text-gray-800"
                  >
                    Blocked
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200">
                <Show when={schedules()}>
                  {(schedules) => {
                    console.log(schedules());
                    return (
                      <For each={schedules()}>
                        {(schedule) => (
                          <tr>
                            <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-800 sm:pl-0">
                              <a
                                href={`/schedules/${schedule.id}`}
                                class="cursor-pointer"
                              >
                                {schedule.name}
                              </a>
                            </td>
                            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                              08:00 am
                            </td>
                            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                              05:00 pm
                            </td>
                            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                              {activeDays(schedule)}
                            </td>
                            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                              {listsAndDomains(schedule)}
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
