# Greyskull LP CLI Implementation Prompts

## Context

We're building a Greyskull LP workout tracking CLI in Go. The application tracks multiple users, their workout programs, and progression over time. Key architectural decisions:

- **Domain Models**: All entities have UUID v7 identifiers for future database migration
- **Storage**: JSON files with case-insensitive username handling
- **Type Safety**: Custom types for LiftName and SetType
- **Programs**: Hardcoded templates with potential for user-created programs later
- **Business Logic**: Modular functions in domain-focused packages

## Prompt 1: Project Foundation & Domain Models

```
I'm building a Greyskull LP workout tracker CLI in Go. Let's start with the foundation and domain models.

Create the initial project structure:
1. Initialize a Go module "github.com/user/greyskull"
2. Add dependencies:
   - github.com/spf13/cobra for CLI
   - github.com/google/uuid for UUID generation (use v7 UUIDs)
   - github.com/stretchr/testify for testing

Create domain models in models/models.go with:

Type definitions:
- type LiftName string
- type SetType string

Constants:
- Squat, Deadlift, BenchPress, OverheadPress (LiftName)
- WarmupSet, WorkingSet, AMRAPSet (SetType)

User domain structs (all with ID field as UUID string):
- User: ID, Username, CurrentProgram (UUID ref), Programs (map[string]*UserProgram), WorkoutHistory ([]Workout), CreatedAt
- UserProgram: ID, UserID, ProgramID, StartingWeights (map[LiftName]float64), CurrentWeights (map[LiftName]float64), CurrentDay (int), StartedAt
- Workout: ID, UserProgramID, Day, Exercises ([]Lift), EnteredAt
- Lift: ID, LiftName, Sets ([]Set)
- Set: ID, Weight, TargetReps, ActualReps (int), Type (SetType), Order (int)

Program template structs:
- Program: ID, Name, Version (string), Workouts ([]WorkoutTemplate), ProgressionRules
- WorkoutTemplate: Day (int), Lifts ([]LiftTemplate)
- LiftTemplate: LiftName, WarmupSets ([]SetTemplate), WorkingSets ([]SetTemplate)
- SetTemplate: Reps (int), WeightPercentage (float64), Type (SetType)
- ProgressionRules: IncreaseRules (map[LiftName]float64), DeloadPercentage (float64), DoubleThreshold (int)

Add validation methods:
- User.Validate() - ensure Username is not empty
- Set.IsComplete() - check if ActualReps > 0

Write comprehensive tests using testify/assert for all validation methods and struct initialization. Ensure UUID generation works correctly.
```

## Prompt 2: Repository Interface & JSON Implementation

```
Now let's create the data persistence layer with proper error handling and thread safety.

Create repository/interface.go with:
- UserRepository interface with methods:
  - Create(user *models.User) error
  - Get(username string) (*models.User, error)
  - Update(user *models.User) error
  - List() ([]string, error)
  - GetCurrent() (string, error)
  - SetCurrent(username string) error

- Sentinel errors:
  - var ErrUserNotFound = errors.New("user not found")
  - var ErrUserAlreadyExists = errors.New("user already exists")
  - var ErrNoCurrentUser = errors.New("no current user set")

Create repository/json.go with JSONUserRepository that:
- Stores user files in os.UserConfigDir()/greyskull/users/{lowercase_username}.json
- Stores current user in os.UserConfigDir()/greyskull/current_user.txt
- Preserves original username casing in User struct while using lowercase for filenames
- Uses sync.Mutex for thread safety
- Creates directories with 0755 permissions and files with 0644
- Implements case-insensitive username lookups (e.g., "Michael" and "michael" are same user)

Constructor should:
- func NewJSONUserRepository() (UserRepository, error)
- Get config directory using os.UserConfigDir()
- Create greyskull/users/ directory structure
- Return error if directory creation fails

Write comprehensive tests using testify/assert and testify/require:
- Test all CRUD operations
- Verify case-insensitive username handling
- Test concurrent access with goroutines
- Use t.TempDir() for test isolation
- Verify error conditions return correct sentinel errors
```

## Prompt 3: Program Templates & Hardcoded Greyskull LP

```
Let's define the program system with our hardcoded Greyskull LP template.

Create program/greyskull_lp.go with:

1. The complete OG Greyskull LP program as a variable:
```go
var GreyskullLP = &models.Program{
    ID:      "550e8400-e29b-41d4-a716-446655440000", // Fixed UUID for consistency
    Name:    "OG Greyskull LP",
    Version: "1.0.0",
    // ... complete structure
}
```

The program should have:

- 6-day workout cycle:
  - Day 1: OverheadPress, Squat
  - Day 2: BenchPress, Deadlift
  - Day 3: OverheadPress, Squat
  - Day 4: BenchPress, Squat
  - Day 5: OverheadPress, Deadlift
  - Day 6: BenchPress, Squat

- Warmup protocol for all lifts (empty bar = 0.0 percentage):
  - 5 reps @ 0.0 (empty bar, 45 lbs)
  - 4 reps @ 0.55 (55% of working weight)
  - 3 reps @ 0.70 (70% of working weight)
  - 2 reps @ 0.85 (85% of working weight)

- Working sets for all lifts:
  - Set 1: 5 reps @ 1.0 (100%), Type: WorkingSet
  - Set 2: 5 reps @ 1.0 (100%), Type: WorkingSet
  - Set 3: 5 reps @ 1.0 (100%), Type: AMRAPSet

- Progression rules:
  - IncreaseRules: OverheadPress/BenchPress = 2.5, Squat/Deadlift = 5.0
  - DeloadPercentage: 0.9
  - DoubleThreshold: 10

2. Program retrieval functions:

- GetByID(id string) (*models.Program, error) - returns GreyskullLP if ID matches
- List() []*models.Program - returns slice containing GreyskullLP

Write tests to verify:

- Program structure matches specification
- All 6 days have correct exercises
- Warmup and working sets are properly configured
- Progression rules are correct

```

## Prompt 4: CLI Foundation & User Commands

```

Let's build the CLI structure with Cobra and implement user management commands.

Create the CLI foundation:

1. main.go:

- Initialize and execute root command
- Handle errors appropriately

2. cmd/root.go:

- Define rootCmd with name "greyskull"
- Set description: "A command-line workout tracker for Greyskull LP"
- Show help when no subcommand provided
- Add Version field (set to "0.1.0" for now)

3. cmd/user.go:

- Create userCmd parent command
- Description: "Manage users"
- Explicitly add child commands in init()

4. cmd/user_create.go:

- Prompt for username using fmt.Print/fmt.Scanln
- Validate username is not empty
- Validate username is filesystem-safe (no special chars like /, \, :, *, ?, ", <, >, |)
- Check for case-insensitive duplicates
- Create User with UUID v7, set CreatedAt
- Initialize empty Programs map
- Save via repository
- Set as current user
- Show success message

5. cmd/user_switch.go:

- Take username as argument
- Validate user exists (case-insensitive lookup)
- Set as current user
- Show confirmation with actual username casing

6. cmd/user_list.go:

- List all users (preserve original casing)
- Mark current user with asterisk (*)
- Show helpful message if no users exist

Wire everything together:

- Root command adds user command
- User command adds create, switch, list commands
- All commands use the JSON repository

Write integration tests:

- Test full user creation flow
- Test case-insensitive operations
- Test switching between users
- Test listing with current user indicator
- Mock user input where needed

```

## Prompt 5: Program Start Command

```

Implement the program start command to initialize a workout program for the current user.

Create cmd/program.go:

- Parent command "program" with description "Manage workout programs"
- Add child commands in init()

Create cmd/program_start.go:

1. Check for current user (error if none)
2. List available programs:
   - Show numbered list (e.g., "1. OG Greyskull LP")
   - Prompt "Select a program (enter number): "
   - Validate selection

3. For each core lift (Squat, Deadlift, Bench Press, Overhead Press):
   - Prompt: "Enter starting weight for {lift} (lbs): "
   - Accept any positive number (float64)
   - Allow decimals for microplates

4. Create UserProgram:
   - Generate UUID v7 for ID
   - Set UserID from current user
   - Set ProgramID from selected program
   - Store starting weights using LiftName keys
   - Copy StartingWeights to CurrentWeights
   - Set CurrentDay to 1
   - Set StartedAt to current time

5. Update user:
   - Add UserProgram to Programs map (keyed by UUID)
   - Set CurrentProgram to the new UserProgram ID
   - Save via repository

6. Show success message with day 1 preview:
   - "Program started! Day 1 will be: {exercises}"

Helper functions:

- promptFloat(prompt string) (float64, error) for weight input
- validatePositive(weight float64) error

Wire to root command.

Write integration tests:

- Mock user input for program selection and weights
- Verify UserProgram created correctly
- Verify CurrentProgram updated
- Verify CurrentDay set to 1
- Test with no current user
- Test invalid weight inputs

```

## Prompt 6: Workout Calculation Engine

```

Create the workout calculation engine that generates workouts from templates.

Create workout/calculator.go with:

1. Weight rounding function:

```go
func RoundDown2_5(weight float64) float64 {
    return math.Floor(weight/2.5) * 2.5
}
```

2. Warmup calculation:

```go
func CalculateWarmupSets(workingWeight float64, warmupTemplates []models.SetTemplate) []models.Set {
    // Return empty slice if working weight < 85 lbs
    // For each template:
    //   - Calculate weight (0.0 = 45 lbs bar, otherwise workingWeight * percentage)
    //   - Round down to 2.5 lbs
    //   - Create Set with UUID, weight, target reps, Type, Order
}
```

3. Working set calculation:

```go
func CalculateWorkingSets(workingWeight float64, workingTemplates []models.SetTemplate) []models.Set {
    // For each template:
    //   - Use working weight directly (can be < 45 lbs)
    //   - Round down to 2.5 lbs
    //   - Create Set with proper Type (WorkingSet or AMRAPSet)
    //   - Set Order field
}
```

4. Main calculation function:

```go
func CalculateNextWorkout(user *models.User, program *models.Program) (*models.Workout, error) {
    // Get current UserProgram from user.CurrentProgram
    // Get current day from UserProgram.CurrentDay
    // Get WorkoutTemplate for that day (handle cycle wrapping)
    // For each LiftTemplate:
    //   - Get current weight from UserProgram.CurrentWeights
    //   - Calculate warmup sets (may be empty)
    //   - Calculate working sets
    //   - Create Lift with all sets
    // Return Workout (don't save it yet, just calculate)
}
```

5. Day cycle helper:

```go
func GetWorkoutDay(currentDay int, totalDays int) int {
    // Handle 1-based indexing and cycling
    // Day 7 should wrap to day 1 for 6-day program
}
```

Write comprehensive tests:

- Test weight rounding (42.7 -> 42.5, 45.0 -> 45.0, etc.)
- Test warmup skipping for weights < 85 lbs
- Test warmup calculation for various weights
- Test working set generation
- Test day cycling (day 7 -> day 1)
- Test full workout calculation
- Verify all Sets have proper Order values

```

## Prompt 7: Next Workout Command

```

Implement the "workout next" command to display the upcoming workout.

Create cmd/workout.go:

- Parent command "workout" with description "Track and view workouts"
- Add child commands in init()

Create cmd/workout_next.go:

1. Load current user:
   - Error if no current user
   - Error if user has no CurrentProgram

2. Get UserProgram and Program:
   - Load UserProgram from user.Programs[user.CurrentProgram]
   - Load Program using program.GetByID(userProgram.ProgramID)

3. Calculate next workout:
   - Use workout.CalculateNextWorkout(user, program)
   - Handle any errors

4. Display workout with clear formatting:

```
Day {N} Workout:
================

Overhead Press:
  Warmup:
    5 reps @ 45 lbs
    4 reps @ 55 lbs
    3 reps @ 70 lbs
    2 reps @ 85 lbs
  Working Sets:
    Set 1: 5 reps @ 95 lbs
    Set 2: 5 reps @ 95 lbs
    Set 3: 5+ reps @ 95 lbs (AMRAP)

Squat:
  Working Sets:
    Set 1: 5 reps @ 135 lbs
    Set 2: 5 reps @ 135 lbs
    Set 3: 5+ reps @ 135 lbs (AMRAP)
```

5. Format helpers:
   - formatWeight(weight float64) string - no decimals for whole numbers
   - formatSet(set models.Set, index int) string
   - Mark AMRAP sets clearly

Wire to root command.

Write integration tests:

- Test with valid program and weights
- Test display for different days
- Test with no current user
- Test with no active program
- Test warmup display vs no warmup (< 85 lbs)
- Verify AMRAP sets marked correctly

```

## Prompt 8: Basic Workout Logging

```

Implement the basic "workout log" command that records completed workouts.

Create cmd/workout_log.go:

1. Command setup:
   - RunE function for the log subcommand
   - No flags initially (--fail comes in next prompt)

2. Load prerequisites:
   - Current user (error if none)
   - UserProgram (error if no active program)
   - Program definition

3. Calculate and display workout:
   - Use workout.CalculateNextWorkout()
   - Show the workout details like "next" command

4. Collect workout data:
   - Auto-complete all warmup sets (ActualReps = TargetReps)
   - Auto-complete working sets (ActualReps = TargetReps)
   - For AMRAP sets only, prompt:
     "How many reps did you complete for {exercise} AMRAP set (5+)? "
   - Validate input is positive integer
   - Store the actual reps

5. Create and save workout:
   - Build models.Workout with UUID, UserProgramID, Day, EnteredAt
   - For each exercise, create models.Lift with completed sets
   - Copy all Set data with ActualReps filled in
   - Add to user.WorkoutHistory
   - Save user via repository

6. Update for next workout:
   - Increment UserProgram.CurrentDay
   - Save user

7. Show completion summary:
   "Workout logged successfully!"
   "Next workout: Day {N}"

Helper functions:

- promptInt(prompt string) (int, error)
- buildCompletedWorkout(template *models.Workout, amrapReps map[models.LiftName]int)*models.Workout

Write tests:

- Test successful logging flow
- Test AMRAP input validation
- Verify workout saved to history
- Verify CurrentDay increments
- Test with no current user/program

```

## Prompt 9: Advanced Workout Logging with Failure Mode

```

Add failure mode support to handle missed reps.

Update cmd/workout_log.go:

1. Add --fail flag:
   - Use cobra.Command.Flags().Bool("fail", false, "Record individual reps for each set")
   - Access with cmd.Flags().GetBool("fail")

2. Modify collection logic:
   - If --fail is false: Keep existing behavior (auto-complete except AMRAP)
   - If --fail is true: Prompt for every set

3. Failure mode prompting:
   For each lift and set:

   ```
   Overhead Press - Set 1 (Warmup):
   Target: 5 reps @ 45 lbs
   How many reps completed?
   ```

   - Accept 0 for failed sets
   - Validate non-negative integer
   - Show set type (Warmup/Working/AMRAP)

4. Update helper functions:
   - collectWithFailure(workout *models.Workout)*models.Workout
   - Shows each set individually
   - Collects actual reps for all sets

5. Keep existing save logic:
   - Same workout history update
   - Same CurrentDay increment

6. Update help text:

   ```
   Log a completed workout. By default, assumes all non-AMRAP sets were completed successfully.
   Use --fail flag to record individual reps for each set.
   ```

Write tests:

- Test default mode still works
- Test --fail flag changes behavior
- Test collecting reps for every set in fail mode
- Test 0 reps accepted for failures
- Verify both modes save correctly
- Test mixed success/failure sets

```

## Prompt 10: Progression System & Final Integration

```

Implement the weight progression system and integrate it with workout logging.

Create workout/progression.go:

1. AMRAP detection:

```go
func GetAMRAPReps(lift *models.Lift) (int, error) {
    // Find the AMRAP set (Type == models.AMRAPSet)
    // Return ActualReps
    // Error if no AMRAP set found
}
```

2. Weight calculation:

```go
func CalculateNewWeight(currentWeight float64, amrapReps int, baseIncrement float64, rules *models.ProgressionRules) float64 {
    // If amrapReps < 5: return currentWeight * rules.DeloadPercentage
    // If amrapReps >= rules.DoubleThreshold: return currentWeight + (baseIncrement * 2)
    // Otherwise: return currentWeight + baseIncrement
    // Round down to 2.5 lbs
}
```

3. Full progression calculation:

```go
func CalculateProgression(workout *models.Workout, currentWeights map[models.LiftName]float64, rules *models.ProgressionRules) (map[models.LiftName]float64, error) {
    // For each lift in workout:
    //   - Get AMRAP reps
    //   - Get base increment from rules.IncreaseRules
    //   - Calculate new weight
    // Return updated weights map
}
```

Update cmd/workout_log.go:

1. After saving workout, before incrementing day:
   - Call progression.CalculateProgression()
   - Update UserProgram.CurrentWeights with new weights
   - Show weight changes to user:

     ```
     Weight Updates:
     Overhead Press: 95 → 97.5 lbs (+2.5)
     Squat: 135 → 140 lbs (+5)
     ```

2. Format weight changes:
   - Show old → new
   - Show difference (+X or -X for deload)
   - Use colors if available (green for increase, red for deload)

Integration updates:

- Ensure CurrentWeights is used for next workout calculation
- Verify CurrentDay increments after progression

Write comprehensive tests:

- Test normal progression (AMRAP = 5-9)
- Test double progression (AMRAP >= 10)
- Test deload (AMRAP < 5)
- Test weight rounding in progression
- Test full integration: log workout → calculate progression → update weights → next workout uses new weights
- Test progression display formatting

Final integration test:

- Create user
- Start program with weights
- Log several workouts with different AMRAP counts
- Verify progression applied correctly
- Verify next workout uses updated weights

```

## Testing Strategy Notes

Each prompt should emphasize:
- Writing tests FIRST (TDD approach)
- Using table-driven tests where appropriate
- Mocking external dependencies
- Testing both happy paths and error conditions
- Using temporary directories for file operations
- Cleaning up test artifacts

## Implementation Notes

- Each prompt builds on the previous implementation
- No orphaned code - everything gets integrated
- Progressive complexity - simple foundations first
- Early testing to catch issues before they compound
- Clear separation of concerns throughout
- Consistent error handling patterns
- User-friendly output at every step## Prompt 8: Basic Workout Logging

```

Implement the basic "workout log" command that records completed workouts.

Create cmd/workout_log.go:

1. Command setup:
   - RunE function for the log subcommand
   - No flags initially (--fail comes in next prompt)

2. Load prerequisites:
   - Current user (error if none)
   - UserProgram (error if no active program)
   - Program definition

3. Calculate and display workout:
   - Use workout.CalculateNextWorkout()
   - Show the workout details like "next" command

4. Collect workout data:
   - Auto-complete all warmup sets (ActualReps = TargetReps)
   - Auto-complete working sets (ActualReps = TargetReps)
   - For AMRAP sets only, prompt:
     "How many reps did you complete for {exercise} AMRAP set (5+)? "
   - Validate input is positive integer
   - Store the actual reps

5. Create and save workout:
   - Build models.Workout with UUID, UserProgramID, Day, EnteredAt
   - For each exercise, create models.Lift with completed sets
   - Copy all Set data with ActualReps filled in
   - Add to user.WorkoutHistory
   - Save user via repository

6. Update for next workout:
   - Increment UserProgram.CurrentDay
   - Save user

7. Show completion summary:
   "Workout logged successfully!"
   "Next workout: Day {N}"

Helper functions:

- promptInt(prompt string) (int, error)
- buildCompletedWorkout(template *models.Workout, amrapReps map[models.LiftName]int)*models.Workout

Write tests:

- Test successful logging flow
- Test AMRAP input validation
- Verify workout saved to history
- Verify CurrentDay increments
- Test with no current user/program

```# Greyskull LP CLI Implementation Prompts

## Prompt 1: Project Foundation & Domain Models

```

I'm building a Greyskull LP workout tracker CLI in Go. Let's start with the foundation and domain models.

Create the initial project structure:

1. Initialize a Go module "github.com/user/greyskull"
2. Add dependencies:
   - github.com/spf13/cobra for CLI
   - github.com/google/uuid for UUID generation (use v7 UUIDs)
   - github.com/stretchr/testify for testing

Create domain models in models/models.go with:

Type definitions:

- type LiftName string
- type SetType string

Constants:

- Squat, Deadlift, BenchPress, OverheadPress (LiftName)
- WarmupSet, WorkingSet, AMRAPSet (SetType)

User domain structs (all with ID field as UUID string):

- User: ID, Username, CurrentProgram (UUID ref), Programs (map[string]*UserProgram), WorkoutHistory ([]Workout), CreatedAt
- UserProgram: ID, UserID, ProgramID, StartingWeights (map[LiftName]float64), CurrentWeights (map[LiftName]float64), CurrentDay (int), StartedAt
- Workout: ID, UserProgramID, Day, Exercises ([]Lift), EnteredAt
- Lift: ID, LiftName, Sets ([]Set)
- Set: ID, Weight, TargetReps, ActualReps (int), Type (SetType), Order (int)

Program template structs:

- Program: ID, Name, Version (string), Workouts ([]WorkoutTemplate), ProgressionRules
- WorkoutTemplate: Day (int), Lifts ([]LiftTemplate)
- LiftTemplate: LiftName, WarmupSets ([]SetTemplate), WorkingSets ([]SetTemplate)
- SetTemplate: Reps (int), WeightPercentage (float64), Type (SetType)
- ProgressionRules: IncreaseRules (map[LiftName]float64), DeloadPercentage (float64), DoubleThreshold (int)

Add validation methods:

- User.Validate() - ensure Username is not empty
- Set.IsComplete() - check if ActualReps > 0

Write comprehensive tests using testify/assert for all validation methods and struct initialization. Ensure UUID generation works correctly.

```

## Prompt 2: Repository Interface & JSON Implementation

```

Now let's create the data persistence layer with proper error handling and thread safety.

Create repository/interface.go with:

- UserRepository interface with methods:
  - Create(user *models.User) error
  - Get(username string) (*models.User, error)
  - Update(user *models.User) error
  - List() ([]string, error)
  - GetCurrent() (string, error)
  - SetCurrent(username string) error

- Sentinel errors:
  - var ErrUserNotFound = errors.New("user not found")
  - var ErrUserAlreadyExists = errors.New("user already exists")
  - var ErrNoCurrentUser = errors.New("no current user set")

Create repository/json.go with JSONUserRepository that:

- Stores user files in os.UserConfigDir()/greyskull/users/{lowercase_username}.json
- Stores current user in os.UserConfigDir()/greyskull/current_user.txt
- Preserves original username casing in User struct while using lowercase for filenames
- Uses sync.Mutex for thread safety
- Creates directories with 0755 permissions and files with 0644
- Implements case-insensitive username lookups (e.g., "Michael" and "michael" are same user)

Constructor should:

- func NewJSONUserRepository() (UserRepository, error)
- Get config directory using os.UserConfigDir()
- Create greyskull/users/ directory structure
- Return error if directory creation fails

Write comprehensive tests using testify/assert and testify/require:

- Test all CRUD operations
- Verify case-insensitive username handling
- Test concurrent access with goroutines
- Use t.TempDir() for test isolation
- Verify error conditions return correct sentinel errors

```

## Prompt 3: Program Templates & Hardcoded Greyskull LP

```

Let's define the program system with our hardcoded Greyskull LP template.

Create program/greyskull_lp.go with:

1. The complete OG Greyskull LP program as a variable:

```go
var GreyskullLP = &models.Program{
    ID:      "550e8400-e29b-41d4-a716-446655440000", // Fixed UUID for consistency
    Name:    "OG Greyskull LP",
    Version: "1.0.0",
    // ... complete structure
}
```

The program should have:

- 6-day workout cycle:
  - Day 1: OverheadPress, Squat
  - Day 2: BenchPress, Deadlift
  - Day 3: OverheadPress, Squat
  - Day 4: BenchPress, Squat
  - Day 5: OverheadPress, Deadlift
  - Day 6: BenchPress, Squat

- Warmup protocol for all lifts (empty bar = 0.0 percentage):
  - 5 reps @ 0.0 (empty bar, 45 lbs)
  - 4 reps @ 0.55 (55% of working weight)
  - 3 reps @ 0.70 (70% of working weight)
  - 2 reps @ 0.85 (85% of working weight)

- Working sets for all lifts:
  - Set 1: 5 reps @ 1.0 (100%), Type: WorkingSet
  - Set 2: 5 reps @ 1.0 (100%), Type: WorkingSet
  - Set 3: 5 reps @ 1.0 (100%), Type: AMRAPSet

- Progression rules:
  - IncreaseRules: OverheadPress/BenchPress = 2.5, Squat/Deadlift = 5.0
  - DeloadPercentage: 0.9
  - DoubleThreshold: 10

2. Program retrieval functions:

- GetByID(id string) (*models.Program, error) - returns GreyskullLP if ID matches
- List() []*models.Program - returns slice containing GreyskullLP

Write tests to verify:

- Program structure matches specification
- All 6 days have correct exercises
- Warmup and working sets are properly configured
- Progression rules are correct

```

## Prompt 4: CLI Foundation & User Commands

```

Let's build the CLI structure with Cobra and implement user management commands.

Create the CLI foundation:

1. main.go:

- Initialize and execute root command
- Handle errors appropriately

2. cmd/root.go:

- Define rootCmd with name "greyskull"
- Set description: "A command-line workout tracker for Greyskull LP"
- Show help when no subcommand provided
- Add Version field (set to "0.1.0" for now)

3. cmd/user.go:

- Create userCmd parent command
- Description: "Manage users"
- Explicitly add child commands in init()

4. cmd/user_create.go:

- Prompt for username using fmt.Print/fmt.Scanln
- Validate username is not empty
- Validate username is filesystem-safe (no special chars like /, \, :, *, ?, ", <, >, |)
- Check for case-insensitive duplicates
- Create User with UUID v7, set CreatedAt
- Initialize empty Programs map
- Save via repository
- Set as current user
- Show success message

5. cmd/user_switch.go:

- Take username as argument
- Validate user exists (case-insensitive lookup)
- Set as current user
- Show confirmation with actual username casing

6. cmd/user_list.go:

- List all users (preserve original casing)
- Mark current user with asterisk (*)
- Show helpful message if no users exist

Wire everything together:

- Root command adds user command
- User command adds create, switch, list commands
- All commands use the JSON repository

Write integration tests:

- Test full user creation flow
- Test case-insensitive operations
- Test switching between users
- Test listing with current user indicator
- Mock user input where needed

```

## Prompt 5: Program Start Command

```

Implement the program start command to initialize a workout program for the current user.

Create cmd/program.go:

- Parent command "program" with description "Manage workout programs"
- Add child commands in init()

Create cmd/program_start.go:

1. Check for current user (error if none)
2. List available programs:
   - Show numbered list (e.g., "1. OG Greyskull LP")
   - Prompt "Select a program (enter number): "
   - Validate selection

3. For each core lift (Squat, Deadlift, Bench Press, Overhead Press):
   - Prompt: "Enter starting weight for {lift} (lbs): "
   - Accept any positive number (float64)
   - Allow decimals for microplates

4. Create UserProgram:
   - Generate UUID v7 for ID
   - Set UserID from current user
   - Set ProgramID from selected program
   - Store starting weights using LiftName keys
   - Copy StartingWeights to CurrentWeights
   - Set CurrentDay to 1
   - Set StartedAt to current time

5. Update user:
   - Add UserProgram to Programs map (keyed by UUID)
   - Set CurrentProgram to the new UserProgram ID
   - Save via repository

6. Show success message with day 1 preview:
   - "Program started! Day 1 will be: {exercises}"

Helper functions:

- promptFloat(prompt string) (float64, error) for weight input
- validatePositive(weight float64) error

Wire to root command.

Write integration tests:

- Mock user input for program selection and weights
- Verify UserProgram created correctly
- Verify CurrentProgram updated
- Verify CurrentDay set to 1
- Test with no current user
- Test invalid weight inputs

```

## Prompt 6: Workout Calculation Engine

```

Create the workout calculation engine that generates workouts from templates.

Create workout/calculator.go with:

1. Weight rounding function:

```go
func RoundDown2_5(weight float64) float64 {
    return math.Floor(weight/2.5) * 2.5
}
```

2. Warmup calculation:

```go
func CalculateWarmupSets(workingWeight float64, warmupTemplates []models.SetTemplate) []models.Set {
    // Return empty slice if working weight < 85 lbs
    // For each template:
    //   - Calculate weight (0.0 = 45 lbs bar, otherwise workingWeight * percentage)
    //   - Round down to 2.5 lbs
    //   - Create Set with UUID, weight, target reps, Type, Order
}
```

3. Working set calculation:

```go
func CalculateWorkingSets(workingWeight float64, workingTemplates []models.SetTemplate) []models.Set {
    // For each template:
    //   - Use working weight directly (can be < 45 lbs)
    //   - Round down to 2.5 lbs
    //   - Create Set with proper Type (WorkingSet or AMRAPSet)
    //   - Set Order field
}
```

4. Main calculation function:

```go
func CalculateNextWorkout(user *models.User, program *models.Program) (*models.Workout, error) {
    // Get current UserProgram from user.CurrentProgram
    // Get current day from UserProgram.CurrentDay
    // Get WorkoutTemplate for that day (handle cycle wrapping)
    // For each LiftTemplate:
    //   - Get current weight from UserProgram.CurrentWeights
    //   - Calculate warmup sets (may be empty)
    //   - Calculate working sets
    //   - Create Lift with all sets
    // Return Workout (don't save it yet, just calculate)
}
```

5. Day cycle helper:

```go
func GetWorkoutDay(currentDay int, totalDays int) int {
    // Handle 1-based indexing and cycling
    // Day 7 should wrap to day 1 for 6-day program
}
```

Write comprehensive tests:

- Test weight rounding (42.7 -> 42.5, 45.0 -> 45.0, etc.)
- Test warmup skipping for weights < 85 lbs
- Test warmup calculation for various weights
- Test working set generation
- Test day cycling (day 7 -> day 1)
- Test full workout calculation
- Verify all Sets have proper Order values

```

## Prompt 7: Next Workout Command

```

Implement the "workout next" command to display the upcoming workout.

Create cmd/workout.go:

- Parent command "workout" with description "Track and view workouts"
- Add child commands in init()

Create cmd/workout_next.go:

1. Load current user:
   - Error if no current user
   - Error if user has no CurrentProgram

2. Get UserProgram and Program:
   - Load UserProgram from user.Programs[user.CurrentProgram]
   - Load Program using program.GetByID(userProgram.ProgramID)

3. Calculate next workout:
   - Use workout.CalculateNextWorkout(user, program)
   - Handle any errors

4. Display workout with clear formatting:

```
Day {N} Workout:
================

Overhead Press:
  Warmup:
    5 reps @ 45 lbs
    4 reps @ 55 lbs
    3 reps @ 70 lbs
    2 reps @ 85 lbs
  Working Sets:
    Set 1: 5 reps @ 95 lbs
    Set 2: 5 reps @ 95 lbs
    Set 3: 5+ reps @ 95 lbs (AMRAP)

Squat:
  Working Sets:
    Set 1: 5 reps @ 135 lbs
    Set 2: 5 reps @ 135 lbs
    Set 3: 5+ reps @ 135 lbs (AMRAP)
```

5. Format helpers:
   - formatWeight(weight float64) string - no decimals for whole numbers
   - formatSet(set models.Set, index int) string
   - Mark AMRAP sets clearly

Wire to root command.

Write integration tests:

- Test with valid program and weights
- Test display for different days
- Test with no current user
- Test with no active program
- Test warmup display vs no warmup (< 85 lbs)
- Verify AMRAP sets marked correctly

```

## Prompt 8: Next Workout Command

```

Implement the "workout next" command to display the upcoming workout.

Create cmd/workout.go with:

1. A parent "workout" command

2. A "next" subcommand that:
   - Loads current user and their active program
   - Uses WorkoutCalculator to determine next workout
   - Displays workout in a clean format:
     - Day number and exercise list
     - For each exercise:
       - Exercise name
       - Warmup sets (e.g., "5 reps @ 45 lbs")
       - Working sets (e.g., "Set 1: 5 reps @ 135 lbs")
       - AMRAP set clearly marked (e.g., "Set 3: 5+ reps @ 135 lbs (AMRAP)")
   - Shows helpful error if no program is active

3. Formatting helpers for:
   - Consistent weight display (no decimals for whole numbers)
   - Clear set numbering
   - Visual separation between exercises

Wire to root command and integrate with repository and calculator.

Include integration tests that:

- Verify correct workout is shown for different days
- Test formatting of output
- Handle case where no program is active
- Verify warmup and working sets display correctly

```

## Prompt 9: Basic Workout Logging

```

Implement the basic "workout log" command (without --fail flag).

Update cmd/workout.go to add:

1. A "log" subcommand that:
   - Gets the current workout using calculator
   - Displays each exercise with its sets
   - Automatically marks warmup sets as complete
   - Automatically marks the 2x5 working sets as complete
   - For the AMRAP set only:
     - Prompts "How many reps did you complete for [exercise] AMRAP set (5+)?"
     - Validates input is a positive integer
     - Accepts the value
   - Creates a CompletedWorkout record with all sets
   - Saves to user's workout history
   - Shows summary of completed workout

2. Helper functions for:
   - Building CompletedWorkout from calculator output
   - Collecting AMRAP reps with validation
   - Displaying completion summary

3. Integration with repository to:
   - Update workout history
   - Save user changes
   - Increment CurrentDay after successful logging

Include tests that:

- Verify warmup and working sets are auto-completed
- Test AMRAP input validation
- Verify workout is saved to history
- Confirm CurrentDay increments
- Test summary display

```

## Prompt 10: Progression System & Final Integration

```

Implement the weight progression system and integrate it with workout logging.

Create workout/progression.go:

1. AMRAP detection:

```go
func GetAMRAPReps(lift *models.Lift) (int, error) {
    // Find the AMRAP set (Type == models.AMRAPSet)
    // Return ActualReps
    // Error if no AMRAP set found
}
```

2. Weight calculation:

```go
func CalculateNewWeight(currentWeight float64, amrapReps int, baseIncrement float64, rules *models.ProgressionRules) float64 {
    // If amrapReps < 5: return currentWeight * rules.DeloadPercentage
    // If amrapReps >= rules.DoubleThreshold: return currentWeight + (baseIncrement * 2)
    // Otherwise: return currentWeight + baseIncrement
    // Round down to 2.5 lbs
}
```

3. Full progression calculation:

```go
func CalculateProgression(workout *models.Workout, currentWeights map[models.LiftName]float64, rules *models.ProgressionRules) (map[models.LiftName]float64, error) {
    // For each lift in workout:
    //   - Get AMRAP reps
    //   - Get base increment from rules.IncreaseRules
    //   - Calculate new weight
    // Return updated weights map
}
```

Update cmd/workout_log.go:

1. After saving workout, before incrementing day:
   - Call progression.CalculateProgression()
   - Update UserProgram.CurrentWeights with new weights
   - Show weight changes to user:

     ```
     Weight Updates:
     Overhead Press: 95 → 97.5 lbs (+2.5)
     Squat: 135 → 140 lbs (+5)
     ```

2. Format weight changes:
   - Show old → new
   - Show difference (+X or -X for deload)
   - Use colors if available (green for increase, red for deload)

Integration updates:

- Ensure CurrentWeights is used for next workout calculation
- Verify CurrentDay increments after progression

Write comprehensive tests:

- Test normal progression (AMRAP = 5-9)
- Test double progression (AMRAP >= 10)
- Test deload (AMRAP < 5)
- Test weight rounding in progression
- Test full integration: log workout → calculate progression → update weights → next workout uses new weights
- Test progression display formatting

Final integration test:

- Create user
- Start program with weights
- Log several workouts with different AMRAP counts
- Verify progression applied correctly
- Verify next workout uses updated weights

```

## Testing Strategy Notes

Each prompt should emphasize:
- Writing tests FIRST (TDD approach)
- Using table-driven tests where appropriate
- Mocking external dependencies
- Testing both happy paths and error conditions
- Using temporary directories for file operations
- Cleaning up test artifacts

## Implementation Notes

- Each prompt builds on the previous implementation
- No orphaned code - everything gets integrated
- Progressive complexity - simple foundations first
- Early testing to catch issues before they compound
- Clear separation of concerns throughout
- Consistent error handling patterns
- User-friendly output at every step

