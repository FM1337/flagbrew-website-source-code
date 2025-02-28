import { POKEMON_LIST } from "../../lib/gpss/pokemon";

export const QUERIES_LIST = [
  {
    name: "Min/Max Level",
    queryField: "min_max_level",
    componentName: "slider-range",
    default: [1, 100],
    value: [],
    operators: [],
    defaultOperator: null,
    operator: null,
    props: {
      solo: true,
      label: "Min/Max level",
      max: 100,
      min: 1,
      "thumb-label": true,
    },
  },
  {
    name: "Generations",
    componentName: "select",
    queryField: "generations",
    default: [1, 2, 3, 4, 5, 6, 7, 8, 9],
    value: [],
    operators: ["IN", "NOT IN"],
    defaultOperator: "IN",
    operator: "IN",
    props: {
      solo: true,
      label: "Generations",
      hint: "Generations to search for",
      rules: [(value) => value.length > 0 || "Required"],
      dense: true,
      items: [1, 2, 3, 4, 5, 6, 7, 8, 9],
      multiple: true,
    },
  },
  {
    name: "Species",
    componentName: "autoComplete",
    // componentName: "select",
    queryField: "species",
    default: [],
    value: [],
    operators: ["IN", "NOT IN"],
    defaultOperator: "IN",
    operator: "IN",
    props: {
      solo: true,
      label: "Species",
      hint: "Species of Pokemon to search for",
      rules: [(value) => value.length > 0 || "Required"],
      autocomplete: true,
      chips: true,
      dense: true,
      "single-line": true,
      "small-chips": true,
      "deletable-chips": true,
      "cache-items": false,
      clearable: true,
      multiple: true,
    },
    update: function(queries, self) {
      // if the user has another generations filter, try to find it.
      var generations = queries.find((q) => {
        return q.queryField == "generations";
      });

      // if user hasn't added a generations filter, just return all.
      if (!generations) {
        self.props.items = POKEMON_LIST.map((p) => p.name);
        return self;
      }

      // filter the species data if they selected a specific set of generations.
      let pokemon = POKEMON_LIST.filter((p) => {
        return generations.value.some((g) => {
          return p.available_generations.indexOf(g) !== -1;
        });
      }).map((p) => p.name);

      // update our properties with the new filtered species.
      self.props.items = pokemon;

      // ensure that if a user already selected a set of species, but then removed the
      // generation that was keeping that species in the list, then remove that species
      // as well, since it can no longer be used.
      for (let pokemon of self.value) {
        if (!self.props.items.includes(pokemon)) {
          self.value.splice(self.value.indexOf(pokemon), 1);
        }
      }

      return self;
    },
  },
  {
    name: "Holding Item",
    componentName: "select",
    queryField: "holding_item",
    default: true,
    value: [],
    operators: ["=", "!="],
    defaultOperator: "=",
    operator: "=",
    props: {
      solo: true,
      label: "Holding Item",
      dense: true,
      items: [
        { text: "Yes", value: true },
        { text: "No", value: false },
      ],
    },
  },
  {
    name: "Nickname",
    componentName: "text",
    queryField: "nickname",
    default: "",
    value: [],
    operators: ["=", "!=", "IN", "NOT IN"],
    defaultOperator: "=",
    operator: "=",
    props: {
      solo: true,
      label: "Nickname",
      rules: [(value) => !!value || "Required"],
      dense: true,
    },
  },
  {
    name: "OT Name",
    componentName: "text",
    queryField: "ot_name",
    default: "",
    value: [],
    operators: ["=", "!=", "IN", "NOT IN"],
    defaultOperator: "=",
    operator: "=",
    props: {
      solo: true,
      label: "Original Trainer Name",
      rules: [(value) => !!value || "Required"],
      dense: true,
    },
  },
  {
    name: "OT ID",
    componentName: "text",
    queryField: "ot_id",
    default: "",
    value: [],
    operators: ["=", "!=", ">", "<", ">=", "<="],
    defaultOperator: "=",
    operator: "=",
    props: {
      solo: true,
      label: "Original Trainer ID",
      rules: [(value) => !!value || "Required"],
      dense: true,
      type: "number",
    },
  },
  {
    name: "HT Name",
    componentName: "text",
    queryField: "ht_name",
    default: "",
    value: [],
    operators: ["=", "!=", "IN", "NOT IN"],
    defaultOperator: "=",
    operator: "=",
    props: {
      solo: true,
      label: "Held Trainer Name",
      rules: [(value) => !!value || "Required"],
      dense: true,
    },
  },
  {
    name: "Legal Only",
    componentName: "select",
    queryField: "legal",
    default: true,
    value: [],
    operators: ["="],
    defaultOperator: "=",
    operator: "=",
    props: {
      solo: true,
      label: "Only List Legal Pokemon",
      dense: true,
      items: [
        { text: "Yes", value: true },
        { text: "No", value: false },
      ],
    },
  },
];
