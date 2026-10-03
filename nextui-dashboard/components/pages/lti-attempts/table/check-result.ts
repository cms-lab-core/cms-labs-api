interface CheckerTaskSummary {
  title: string;
  complete: boolean;
}

export interface AttemptCheckSummary {
  currentScore?: number;
  maxScore?: number;
  resultDisplay?: string;
  tasks: CheckerTaskSummary[];
  passedTasks: number;
  progress: number;
}

type UnknownRecord = Record<string, unknown>;

const isRecord = (value: unknown): value is UnknownRecord =>
  typeof value === 'object' && value !== null && !Array.isArray(value);

const readValue = (record: UnknownRecord, camelCase: string, snakeCase: string): unknown =>
  record[camelCase] ?? record[snakeCase];

const readNumber = (value: unknown): number | undefined => {
  if (typeof value === 'number' && Number.isFinite(value)) return value;
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) return parsed;
  }
  return undefined;
};

const readString = (value: unknown): string | undefined =>
  typeof value === 'string' && value.trim() !== '' ? value.trim() : undefined;

export const parseAttemptCheckResult = (value: unknown): AttemptCheckSummary | undefined => {
  if (!isRecord(value)) return undefined;

  const tasksValue = readValue(value, 'tasks', 'tasks');
  const tasks = Array.isArray(tasksValue)
    ? tasksValue.filter(isRecord).map((task, index) => ({
        title: readString(readValue(task, 'title', 'title')) || `#${index + 1}`,
        complete: readValue(task, 'complete', 'complete') === true
      }))
    : [];
  const currentScore = readNumber(readValue(value, 'currentScore', 'current_score'));
  const maxScore = readNumber(readValue(value, 'maxScore', 'max_score'));
  const resultDisplay = readString(readValue(value, 'resultDisplay', 'result_display'));
  const checkID = readString(readValue(value, 'checkId', 'check_id'));
  const report = readString(readValue(value, 'report', 'report'));
  const logs = readString(readValue(value, 'logs', 'logs'));

  if (
    !checkID &&
    !resultDisplay &&
    !report &&
    !logs &&
    tasks.length === 0 &&
    currentScore === undefined &&
    maxScore === undefined
  ) {
    return undefined;
  }

  const passedTasks = tasks.filter((task) => task.complete).length;
  const progress =
    tasks.length > 0
      ? Math.round((passedTasks / tasks.length) * 100)
      : maxScore && maxScore > 0 && currentScore !== undefined
        ? Math.min(100, Math.max(0, Math.round((currentScore / maxScore) * 100)))
        : 100;

  return { currentScore, maxScore, resultDisplay, tasks, passedTasks, progress };
};

export const formatCheckScore = (value: number): string =>
  new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(value);
