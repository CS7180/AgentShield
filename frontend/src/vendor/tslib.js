export function __awaiter(thisArg, _arguments, P, generator) {
  function adopt(value) {
    return value instanceof P ? value : new P((resolve) => resolve(value));
  }

  return new (P || Promise)((resolve, reject) => {
    function fulfilled(value) {
      try {
        step(generator.next(value));
      } catch (error) {
        reject(error);
      }
    }

    function rejected(value) {
      try {
        step(generator.throw(value));
      } catch (error) {
        reject(error);
      }
    }

    function step(result) {
      if (result.done) {
        resolve(result.value);
        return;
      }

      adopt(result.value).then(fulfilled, rejected);
    }

    step((generator = generator.apply(thisArg, _arguments || [])).next());
  });
}

export function __rest(source, excluded) {
  const target = {};

  for (const key in source) {
    if (Object.prototype.hasOwnProperty.call(source, key) && !excluded.includes(key)) {
      target[key] = source[key];
    }
  }

  if (source != null && typeof Object.getOwnPropertySymbols === 'function') {
    for (const symbol of Object.getOwnPropertySymbols(source)) {
      if (
        !excluded.includes(symbol) &&
        Object.prototype.propertyIsEnumerable.call(source, symbol)
      ) {
        target[symbol] = source[symbol];
      }
    }
  }

  return target;
}
