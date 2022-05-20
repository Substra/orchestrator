# Algo

An Algo is a structure which represents a function, with inputs and outputs.

## Algo inputs/outputs
### Structure

An Algo input/output has:

- **An `identifier`**
  - It identifies inputs (or outputs, respectively) from each other.
  - Identifiers are used in the definition of [Compute Tasks](./computetask.md). A compute task's input identifiers must match the Algo input identifiers.
  - Output identifiers can be used to reference a compute task output as the input of another task.
- **An asset `kind`**
  - This is the kind of asset for an Algo input/output. For example, an input of kind `DATA_SAMPLE` will represent input data samples.
- And a set of options (see below)
  - **`Multiple`**: The input/output can have multiple values
  - **`Optional`** (input only): The input is optional

### Constraints and validation

All input `identifier`s must be between 0 and 100 characters.

#### Inputs

Algo inputs must verify the following constraints:

- An input `kind` must be one of the following: `MODEL`, `DATA_SAMPLE`, `DATA_MANAGER`
- An input of kind `DATA_MANAGER` cannot be `Optional` nor `Multiple`
- It is not allowed to have multiple inputs of kind `DATA_MANAGER`
- If there is an input of kind `DATA_MANAGER`, there must be an input of kind `DATA_SAMPLES`, and vice versa

#### Outputs

Algo outputs must verify the following constraints:

- An output `kind` must be one of the following: `MODEL`, `PERFORMANCE`
- An output of kind `PERFORMANCE` cannot be `Multiple`
