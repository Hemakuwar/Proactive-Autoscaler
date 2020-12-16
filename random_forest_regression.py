import sys
import math
from json import JSONDecodeError
from datetime import datetime, timedelta
from dataclasses import dataclass
from typing import List, Optional
from sklearn.ensemble import RandomForestRegressor
from dataclasses_json import dataclass_json, LetterCase



@dataclass_json(letter_case=LetterCase.CAMEL)
@dataclass
class EvaluationValue:

    target_replicas: int

@dataclass_json(letter_case=LetterCase.CAMEL)
@dataclass
class Evaluation:

    id: int
    created: str
    val: EvaluationValue

@dataclass_json(letter_case=LetterCase.CAMEL)
@dataclass
class AlgorithmInput:

    look_ahead: int
    evaluations: List[Evaluation]
    current_time: Optional[str] = None

stdin = sys.stdin.read()

if stdin is None or stdin == "":
    print("No standard input provided to Random Forest Regression algorithm, exiting", file=sys.stderr)
    sys.exit(1)

try:
    algorithm_input = AlgorithmInput.from_json(stdin)
except JSONDecodeError as ex:
    print("Invalid JSON provided: {0}, exiting".format(str(ex)), file=sys.stderr)
    sys.exit(1)
except KeyError as ex:
    print("Invalid JSON provided: missing {0}, exiting".format(str(ex)), file=sys.stderr)
    sys.exit(1)

current_time = datetime.utcnow()

if algorithm_input.current_time is not None:
    try:
        current_time = datetime.strptime(algorithm_input.current_time, "%Y-%m-%dT%H:%M:%SZ")
    except ValueError as ex:
        print("Invalid datetime format: {0}".format(str(ex)), file=sys.stderr)
        sys.exit(1)

search_time = datetime.timestamp(current_time + timedelta(milliseconds=int(algorithm_input.look_ahead)))

a = []
b = []


for i, evaluation in enumerate(algorithm_input.evaluations):
    try:
        created = datetime.strptime(evaluation.created, "%Y-%m-%dT%H:%M:%SZ")
    except ValueError as ex:
        print("Invalid datetime format: {0}".format(str(ex)), file=sys.stderr)
        sys.exit(1)

    a.append(search_time - datetime.timestamp(created))
    b.append(evaluation.val.target_replicas)
    
x = a.x.values.reshape(-1, 1)
y = b.values.reshape(-1, 1)
x_train, x_test, y_train, y_test = train_test_split(x, y, test_size=0.30, random_state=42)

regr = RandomForestRegressor(max_depth=4, random_state=0)
regr.fit(x_train, y_train)

print(math.ceil(regr.predict(x_test)))
