import { apiFetch } from "./api.js";
document.addEventListener("DOMContentLoaded", async function () {
  const tasksContainer = document.getElementById("tasks-container");
  const addTaskForm = document.getElementById("addTaskForm");
  const unassignedTasksContainer = document.getElementById(
    "unassigned-tasks-container",
  );

  let draggedTaskId = null;
  let draggedTaskOwner = null;
  let draggedTaskStatus = null;
  let selectedTaskId = null;

  const statusTranslations = {
    "To Do": "待办",
    "In Progress": "进行中",
    Stalled: "已停滞",
    Done: "已完成",
  };

  async function fetchTasks() {
    try {
      const tasks = await apiFetch("/api/tasks");
      const bugs = await apiFetch("/api/bugs");
      renderTasks(tasks || [], bugs || {});
      await fetchAllHistory();
    } catch (error) {
      console.error("Error fetching tasks:", error);
    }
  }

  function renderTasks(tasks, bugs) {
    tasksContainer.innerHTML = "";
    unassignedTasksContainer.innerHTML = "";

    const unassignedTasks = tasks.filter((task) => !task.owner);
    const assignedTasks = tasks.filter((task) => task.owner);

    unassignedTasks.forEach((task) => {
      const taskElement = createTaskElement(task);
      unassignedTasksContainer.appendChild(taskElement);
    });

    unassignedTasksContainer.addEventListener("dragover", dragOver);
    unassignedTasksContainer.addEventListener("dragleave", dragLeave);
    unassignedTasksContainer.addEventListener("drop", dropTask);

    const groupedByOwnerAndStatus = assignedTasks.reduce((acc, task) => {
      if (!acc[task.owner]) {
        acc[task.owner] = {};
      }
      if (!acc[task.owner][task.status]) {
        acc[task.owner][task.status] = [];
      }
      acc[task.owner][task.status].push(task);
      return acc;
    }, {});

    const allOwners = Array.from(
      new Set(assignedTasks.map((task) => task.owner)),
    ).sort();
    const statuses = ["In Progress", "Stalled", "To Do", "Done"];

    const kanbanTable = document.createElement("div");
    kanbanTable.className = "kanban-table";

    const headerCorner = document.createElement("div");
    headerCorner.className = "kanban-table-header";
    kanbanTable.appendChild(headerCorner);

    statuses.forEach((status) => {
      const headerCell = document.createElement("div");
      const statusSlug = status.toLowerCase().replace(" ", "-");
      headerCell.className = `kanban-table-header status-header-${statusSlug}`;
      headerCell.textContent = statusTranslations[status];
      kanbanTable.appendChild(headerCell);
    });

    allOwners.forEach((owner) => {
      const ownerCell = document.createElement("div");
      ownerCell.className = "kanban-table-cell kanban-table-owner-cell";
      console.log(bugs);
      // const ownerBugs = bugs[owner] || 0;
      const solvedBugs = bugs["resolved"][owner] || 0;
      const unsolvedBugs = bugs["unresolved"][owner] || 0;
      ownerCell.innerHTML = `
                <span>${owner}</span>
                <div class="resolved-bugs-count" style="bottom: 20px;">待解决 bug: <span style="font-family: monospace; display: inline-block; width: 3ch; text-align: right;">${unsolvedBugs}</span></div>
                <div class="resolved-bugs-count">已解决 bug: <span style="font-family: monospace; display: inline-block; width: 3ch; text-align: right;">${solvedBugs}</span></div><br>
            `;
      kanbanTable.appendChild(ownerCell);

      statuses.forEach((status) => {
        const column = document.createElement("div");
        const statusSlug = status.toLowerCase().replace(" ", "-");
        column.className = `kanban-table-cell kanban-table-status-column status-${statusSlug}`;
        column.setAttribute("data-status", status);
        column.setAttribute("data-owner", owner);
        column.addEventListener("dragover", dragOver);
        column.addEventListener("dragleave", dragLeave);
        column.addEventListener("drop", dropTask);

        const tasksForStatus =
          (groupedByOwnerAndStatus[owner] &&
            groupedByOwnerAndStatus[owner][status]) ||
          [];
        tasksForStatus.sort((a, b) => {
          if (!a.endTime) return 1;
          if (!b.endTime) return -1;
          return new Date(b.endTime) - new Date(a.endTime);
        });

        if (status === "Done" && tasksForStatus.length > 5) {
          const tasksToShow = tasksForStatus.slice(0, 5);
          tasksToShow.forEach((task) => {
            const taskElement = createTaskElement(task);
            column.appendChild(taskElement);
          });

          const showMoreContainer = document.createElement("div");
          showMoreContainer.className = "show-more-container";

          const showMoreButton = document.createElement("button");
          showMoreButton.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg>`;
          showMoreButton.className = "show-more-button";

          const showMoreCount = document.createElement("span");
          showMoreCount.className = "show-more-count";
          showMoreCount.textContent = `${tasksForStatus.length - 5}`;

          showMoreContainer.appendChild(showMoreButton);
          showMoreContainer.appendChild(showMoreCount);
          column.appendChild(showMoreContainer);

          showMoreContainer.addEventListener("click", () => {
            column.innerHTML = ""; // Clear the column
            tasksForStatus.forEach((task) => {
              const taskElement = createTaskElement(task);
              column.appendChild(taskElement);
            });
          });
        } else {
          tasksForStatus.forEach((task) => {
            const taskElement = createTaskElement(task);
            column.appendChild(taskElement);
          });
        }
        kanbanTable.appendChild(column);
      });
    });
    tasksContainer.appendChild(kanbanTable);
  }

  function createTaskElement(task) {
    const taskElement = document.createElement("div");
    taskElement.className = "task";
    taskElement.setAttribute("draggable", "true");
    taskElement.setAttribute("data-id", task.id);
    taskElement.setAttribute("data-owner", task.owner);
    taskElement.setAttribute("data-status", task.status);
    taskElement.addEventListener("dragstart", dragStart);
    taskElement.addEventListener("click", selectTask);

    taskElement.innerHTML = `
            <div class="details">
                <h4 class="task-title-with-dates">
                    ${task.title}
                    <span class="task-date-display">
                        <input type="date" class="start-date-input" data-id="${task.id}" value="${task.startTime || ""}">
                        ~
                        <input type="date" class="end-date-input" data-id="${task.id}" value="${task.endTime || ""}">
                    </span>
                </h4>
            </div>
        `;
    return taskElement;
  }

  function dragStart(e) {
    draggedTaskId = e.target.dataset.id;
    draggedTaskOwner = e.target.dataset.owner;
    draggedTaskStatus = e.target.dataset.status;
    e.dataTransfer.setData("text/plain", draggedTaskId);
    e.target.classList.add("dragging");
  }

  function dragOver(e) {
    e.preventDefault();
    const target = e.target.closest(
      ".kanban-table-status-column, #unassigned-tasks-container",
    );
    if (target) {
      target.classList.add("drag-over");
    }
  }

  function dragLeave(e) {
    const target = e.target.closest(
      ".kanban-table-status-column, #unassigned-tasks-container",
    );
    if (target) {
      target.classList.remove("drag-over");
    }
  }

  async function dropTask(e) {
    e.preventDefault();
    const targetColumn = e.target.closest(
      ".kanban-table-status-column, #unassigned-tasks-container",
    );
    if (!targetColumn) return;

    targetColumn.classList.remove("drag-over");

    const newStatus = targetColumn.dataset.status || "To Do";
    const newOwner = targetColumn.dataset.owner || "";

    if (
      draggedTaskId &&
      (draggedTaskOwner !== newOwner || draggedTaskStatus !== newStatus)
    ) {
      try {
        await apiFetch(`/api/tasks/${draggedTaskId}`, {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ status: newStatus, owner: newOwner }),
        });
        await fetchTasks();
      } catch (error) {
        console.error("Error updating task:", error);
      }
    }

    const draggedElement = document.querySelector(
      `.task[data-id="${draggedTaskId}"]`,
    );
    if (draggedElement) {
      draggedElement.classList.remove("dragging");
    }
    draggedTaskId = null;
    draggedTaskOwner = null;
    draggedTaskStatus = null;
  }

  function deselectCurrentTask() {
    if (selectedTaskId) {
      const prevSelectedTask = document.querySelector(
        `.task[data-id="${selectedTaskId}"]`,
      );
      if (prevSelectedTask) {
        prevSelectedTask.classList.remove("selected");
      }
      selectedTaskId = null;
    }
  }

  function selectTask(e) {
    const clickedTask = e.currentTarget;
    const taskId = clickedTask.dataset.id;

    if (selectedTaskId === taskId) {
      deselectCurrentTask();
    } else {
      deselectCurrentTask();
      clickedTask.classList.add("selected");
      selectedTaskId = taskId;
    }
  }

  document.addEventListener("click", function (e) {
    if (!e.target.closest(".task")) {
      deselectCurrentTask();
    }
  });

  document.addEventListener("keydown", async function (e) {
    if (e.key === "Backspace" && selectedTaskId) {
      e.preventDefault();
      if (confirm("确定要删除此任务吗？")) {
        try {
          await apiFetch(`/api/tasks/${selectedTaskId}`, {
            method: "DELETE",
          });
          selectedTaskId = null;
          await fetchTasks();
        } catch (error) {
          console.error("Error deleting task:", error);
        }
      }
    }
  });

  addTaskForm.addEventListener("submit", async function (event) {
    event.preventDefault();
    const title = document.getElementById("taskTitle").value;
    const owner = document.getElementById("taskOwner").value.trim();
    try {
      await apiFetch("/api/tasks", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ title, owner: owner, status: "To Do" }),
      });
      await fetchTasks();
      addTaskForm.reset();
    } catch (error) {
      console.error("Error adding task:", error);
    }
  });

  tasksContainer.addEventListener("change", async function (event) {
    const target = event.target;
    if (
      target.classList.contains("end-date-input") ||
      target.classList.contains("start-date-input")
    ) {
      const id = target.dataset.id;
      const key = target.classList.contains("end-date-input")
        ? "endTime"
        : "startTime";
      const value = target.value;

      try {
        await apiFetch(`/api/tasks/${id}`, {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ [key]: value }),
        });
      } catch (error) {
        console.error("Error updating task date:", error);
      }
    }
  });

  async function fetchAllHistory() {
    try {
      const history = await apiFetch("/api/history");
      const globalHistoryContent = document.getElementById(
        "globalHistoryContent",
      );
      globalHistoryContent.innerHTML = "";
      if (history && history.length > 0) {
        history.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
        history.forEach((record) => {
          const recordElement = document.createElement("p");
          const timestamp = new Date(record.timestamp).toLocaleString();
          recordElement.textContent = `${timestamp}: 任务 [${record.taskTitle} (${record.taskOwner})] 的字段 '${record.field}' 从 '${record.oldValue}' 变为 '${record.newValue}'`;
          globalHistoryContent.appendChild(recordElement);
        });
      } else {
        globalHistoryContent.textContent = "没有历史记录。";
      }
    } catch (error) {
      console.error("Error fetching history:", error);
    }
  }

  await fetchTasks();
});
