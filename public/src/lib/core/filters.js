import Vue from "vue";
import { DateTime } from "luxon";

export const dateFilter = Vue.filter("date", (v, includeTime) => {
  return DateTime.fromISO(v).toLocaleString(includeTime ? DateTime.DATETIME_MED : DateTime.DATE_MED_WITH_WEEKDAY);
});
