# Greyskull LP CLI Implementation Todo Checklist

## Prompt 1: Project Foundation & Domain Models

### Project Setup
- [x] Initialize Go module "github.com/mikowitz/greyskull"
- [x] Add dependency: github.com/spf13/cobra for CLI
- [x] Add dependency: github.com/google/uuid for UUID generation (use v7 UUIDs)
- [x] Add dependency: github.com/stretchr/testify for testing

### Domain Models (models/models.go)
- [x] Create type definitions:
  - [x] `type LiftName string`
  - [x] `type SetType string`

- [x] Create constants:
  - [x] LiftName constants: Squat, Deadlift, BenchPress, OverheadPress
  - [x] SetType constants: WarmupSet, WorkingSet, AMRAPSet

- [x] Create User domain structs (all with UUID ID):
  - [x] User: ID, Username, CurrentProgram (UUID ref), Programs (map[uuid.UUID]*UserProgram), WorkoutHistory ([]Workout), CreatedAt
  - [x] UserProgram: ID, UserID, ProgramID, StartingWeights (map[LiftName]float64), CurrentWeights (map[LiftName]float64), CurrentDay (int), StartedAt
  - [x] Workout: ID, UserProgramID, Day, Exercises ([]Lift), EnteredAt
  - [x] Lift: ID, LiftName, Sets ([]Set)
  - [x] Set: ID, Weight, TargetReps, ActualReps (int), Type (SetType), Order (int)

- [x] Create Program template structs:
  - [x] Program: ID, Name, Version (string), Workouts ([]WorkoutTemplate), ProgressionRules
  - [x] WorkoutTemplate: Day (int), Lifts ([]LiftTemplate)
  - [x] LiftTemplate: LiftName, WarmupSets ([]SetTemplate), WorkingSets ([]SetTemplate)
  - [x] SetTemplate: Reps (int), WeightPercentage (float64), Type (SetType)
  - [x] ProgressionRules: IncreaseRules (map[LiftName]float64), DeloadPercentage (float64), DoubleThreshold (int)

- [x] Add validation methods:
  - [x] User.Validate() - ensure Username starts with letter and contains only letters, numbers, and dashes
  - [x] Set.IsComplete() - check if ActualReps > 0

### Testing
- [x] Write comprehensive tests using testify/assert for all validation methods
- [x] Test struct initialization
- [x] Ensure UUID generation works correctly

## Prompt 2: Repository Interface & JSON Implementation

### Repository Interface (repository/interface.go)
- [x] Create UserRepository interface with methods:
  - [x] Create(user *models.User) error
  - [x] Get(username string) (*models.User, error)
  - [x] Update(user *models.User) error
  - [x] List() ([]string, error)
  - [x] GetCurrent() (string, error)
  - [x] SetCurrent(username string) error

- [x] Create sentinel errors:
  - [x] var ErrUserNotFound = errors.New("user not found")
  - [x] var ErrUserAlreadyExists = errors.New("user already exists")
  - [x] var ErrNoCurrentUser = errors.New("no current user set")

### JSON Repository Implementation (repository/json.go)
- [x] Create JSONUserRepository that:
  - [x] Stores user files in os.UserConfigDir()/greyskull/users/{lowercase_username}.json
  - [x] Stores current user in os.UserConfigDir()/greyskull/current_user.txt
  - [x] Preserves original username casing in User struct while using lowercase for filenames
  - [x] Uses sync.Mutex for thread safety
  - [x] Creates directories with 0755 permissions and files with 0644
  - [x] Implements case-insensitive username lookups

- [x] Create constructor:
  - [x] func NewJSONUserRepository() (UserRepository, error)
  - [x] Get config directory using os.UserConfigDir()
  - [x] Create greyskull/users/ directory structure
  - [x] Return error if directory creation fails

### Testing
- [x] Test all CRUD operations
- [x] Verify case-insensitive username handling
- [x] Test concurrent access with goroutines
- [x] Use t.TempDir() for test isolation
- [x] Verify error conditions return correct sentinel errors

## Prompt 3: Program Templates & Hardcoded Greyskull LP

### Program Definition (program/greyskull_lp.go)
- [x] Create complete OG Greyskull LP program variable:
  - [x] ID: "550e8400-e29b-41d4-a716-446655440000" (Fixed UUID)
  - [x] Name: "OG Greyskull LP"
  - [x] Version: "1.0.0"

- [x] Define 6-day workout cycle:
  - [x] Day 1: OverheadPress, Squat
  - [x] Day 2: BenchPress, Deadlift
  - [x] Day 3: OverheadPress, Squat
  - [x] Day 4: BenchPress, Squat
  - [x] Day 5: OverheadPress, Deadlift
  - [x] Day 6: BenchPress, Squat

- [x] Define warmup protocol for all lifts:
  - [x] 5 reps @ 0.0 (empty bar, 45 lbs)
  - [x] 4 reps @ 0.55 (55% of working weight)
  - [x] 3 reps @ 0.70 (70% of working weight)
  - [x] 2 reps @ 0.85 (85% of working weight)

- [x] Define working sets for all lifts:
  - [x] Set 1: 5 reps @ 1.0 (100%), Type: WorkingSet
  - [x] Set 2: 5 reps @ 1.0 (100%), Type: WorkingSet
  - [x] Set 3: 5 reps @ 1.0 (100%), Type: AMRAPSet

- [x] Define progression rules:
  - [x] IncreaseRules: OverheadPress/BenchPress = 2.5, Squat/Deadlift = 5.0
  - [x] DeloadPercentage: 0.9
  - [x] DoubleThreshold: 10

### Program Functions
- [x] GetByID(id string) (*models.Program, error) - returns GreyskullLP if ID matches
- [x] List() []*models.Program - returns slice containing GreyskullLP

### Testing
- [x] Verify program structure matches specification
- [x] Test all 6 days have correct exercises
- [x] Verify warmup and working sets are properly configured
- [x] Verify progression rules are correct

## Prompt 4: CLI Foundation & User Commands

### CLI Foundation
- [x] Create main.go:
  - [x] Initialize and execute root command
  - [x] Handle errors appropriately

- [x] Create cmd/root.go:
  - [x] Define rootCmd with name "greyskull"
  - [x] Set description: "A command-line workout tracker for Greyskull LP"
  - [x] Show help when no subcommand provided
  - [x] Add Version field (set to "0.1.0")

### User Commands
- [x] Create cmd/user.go:
  - [x] Create userCmd parent command
  - [x] Description: "Manage users"
  - [x] Explicitly add child commands in init()

- [x] Create cmd/user_create.go:
  - [x] Prompt for username using fmt.Print/fmt.Scanln
  - [x] Validate username is not empty
  - [x] Validate username is filesystem-safe (no special chars)
  - [x] Check for case-insensitive duplicates
  - [x] Create User with UUID v7, set CreatedAt
  - [x] Initialize empty Programs map
  - [x] Save via repository
  - [x] Set as current user
  - [x] Show success message

- [x] Create cmd/user_switch.go:
  - [x] Take username as argument
  - [x] Validate user exists (case-insensitive lookup)
  - [x] Set as current user
  - [x] Show confirmation with actual username casing

- [x] Create cmd/user_list.go:
  - [x] List all users (preserve original casing)
  - [x] Mark current user with asterisk (*)
  - [x] Show helpful message if no users exist

### Integration
- [x] Wire root command to add user command
- [x] Wire user command to add create, switch, list commands
- [x] Ensure all commands use the JSON repository
- [x] Fix XDG_CONFIG_HOME support for proper test isolation
- [x] Fix command output to use cmd.OutOrStdout() for test capturing

### Testing
- [x] Test full user creation flow
- [x] Test case-insensitive operations
- [x] Test switching between users
- [x] Test listing with current user indicator
- [x] Write integration tests avoiding complex stdin mocking
- [x] Test username validation comprehensively
- [x] Test help text for all commands

## Prompt 5: Program Start Command

### Program Commands
- [ ] Create cmd/program.go:
  - [ ] Parent command "program" with description "Manage workout programs"
  - [ ] Add child commands in init()

- [ ] Create cmd/program_start.go:
  - [ ] Check for current user (error if none)
  - [ ] List available programs with numbered list
  - [ ] Prompt "Select a program (enter number): "
  - [ ] Validate selection

### Weight Input
- [ ] For each core lift (Squat, Deadlift, Bench Press, Overhead Press):
  - [ ] Prompt: "Enter starting weight for {lift} (lbs): "
  - [ ] Accept any positive number (float64)
  - [ ] Allow decimals for microplates

### UserProgram Creation
- [ ] Create UserProgram:
  - [ ] Generate UUID v7 for ID
  - [ ] Set UserID from current user
  - [ ] Set ProgramID from selected program
  - [ ] Store starting weights using LiftName keys
  - [ ] Copy StartingWeights to CurrentWeights
  - [ ] Set CurrentDay to 1
  - [ ] Set StartedAt to current time

### User Update
- [ ] Update user:
  - [ ] Add UserProgram to Programs map (keyed by UUID)
  - [ ] Set CurrentProgram to the new UserProgram ID
  - [ ] Save via repository

- [ ] Show success message with day 1 preview

### Helper Functions
- [ ] promptFloat(prompt string) (float64, error) for weight input
- [ ] validatePositive(weight float64) error

### Integration & Testing
- [ ] Wire to root command
- [ ] Mock user input for program selection and weights
- [ ] Verify UserProgram created correctly
- [ ] Verify CurrentProgram updated
- [ ] Verify CurrentDay set to 1
- [ ] Test with no current user
- [ ] Test invalid weight inputs

## Prompt 6: Workout Calculation Engine

### Calculator Functions (workout/calculator.go)
- [ ] Create weight rounding function:
  - [ ] func RoundDown2_5(weight float64) float64

- [ ] Create warmup calculation:
  - [ ] func CalculateWarmupSets(workingWeight float64, warmupTemplates []models.SetTemplate) []models.Set
  - [ ] Return empty slice if working weight < 85 lbs
  - [ ] Calculate weight (0.0 = 45 lbs bar, otherwise workingWeight * percentage)
  - [ ] Round down to 2.5 lbs
  - [ ] Create Set with UUID, weight, target reps, Type, Order

- [ ] Create working set calculation:
  - [ ] func CalculateWorkingSets(workingWeight float64, workingTemplates []models.SetTemplate) []models.Set
  - [ ] Use working weight directly (can be < 45 lbs)
  - [ ] Round down to 2.5 lbs
  - [ ] Create Set with proper Type (WorkingSet or AMRAPSet)
  - [ ] Set Order field

- [ ] Create main calculation function:
  - [ ] func CalculateNextWorkout(user *models.User, program *models.Program) (*models.Workout, error)
  - [ ] Get current UserProgram from user.CurrentProgram
  - [ ] Get current day from UserProgram.CurrentDay
  - [ ] Get WorkoutTemplate for that day (handle cycle wrapping)
  - [ ] For each LiftTemplate: calculate warmup and working sets
  - [ ] Create Lift with all sets
  - [ ] Return Workout (don't save it yet, just calculate)

- [ ] Create day cycle helper:
  - [ ] func GetWorkoutDay(currentDay int, totalDays int) int
  - [ ] Handle 1-based indexing and cycling
  - [ ] Day 7 should wrap to day 1 for 6-day program

### Testing
- [ ] Test weight rounding (42.7 -> 42.5, 45.0 -> 45.0, etc.)
- [ ] Test warmup skipping for weights < 85 lbs
- [ ] Test warmup calculation for various weights
- [ ] Test working set generation
- [ ] Test day cycling (day 7 -> day 1)
- [ ] Test full workout calculation
- [ ] Verify all Sets have proper Order values

## Prompt 7: Next Workout Command

### Workout Commands
- [ ] Create cmd/workout.go:
  - [ ] Parent command "workout" with description "Track and view workouts"
  - [ ] Add child commands in init()

- [ ] Create cmd/workout_next.go:
  - [ ] Load current user (error if no current user)
  - [ ] Error if user has no CurrentProgram
  - [ ] Load UserProgram from user.Programs[user.CurrentProgram]
  - [ ] Load Program using program.GetByID(userProgram.ProgramID)
  - [ ] Calculate next workout using workout.CalculateNextWorkout(user, program)
  - [ ] Handle any errors

### Display Formatting
- [ ] Display workout with clear formatting:
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
  ```

### Format Helpers
- [ ] formatWeight(weight float64) string - no decimals for whole numbers
- [ ] formatSet(set models.Set, index int) string
- [ ] Mark AMRAP sets clearly

### Integration & Testing
- [ ] Wire to root command
- [ ] Test with valid program and weights
- [ ] Test display for different days
- [ ] Test with no current user
- [ ] Test with no active program
- [ ] Test warmup display vs no warmup (< 85 lbs)
- [ ] Verify AMRAP sets marked correctly

## Prompt 8: Basic Workout Logging

### Command Setup (cmd/workout_log.go)
- [ ] RunE function for the log subcommand
- [ ] No flags initially (--fail comes in next prompt)

### Prerequisites Loading
- [ ] Check for current user (error if none)
- [ ] Load UserProgram (error if no active program)
- [ ] Load Program definition

### Workout Display & Collection
- [ ] Calculate and display workout using workout.CalculateNextWorkout()
- [ ] Show the workout details like "next" command
- [ ] Auto-complete all warmup sets (ActualReps = TargetReps)
- [ ] Auto-complete working sets (ActualReps = TargetReps)
- [ ] For AMRAP sets only, prompt: "How many reps did you complete for {exercise} AMRAP set (5+)? "
- [ ] Validate input is positive integer
- [ ] Store the actual reps

### Workout Creation & Saving
- [ ] Build models.Workout with UUID, UserProgramID, Day, EnteredAt
- [ ] For each exercise, create models.Lift with completed sets
- [ ] Copy all Set data with ActualReps filled in
- [ ] Add to user.WorkoutHistory
- [ ] Save user via repository

### Progression
- [ ] Increment UserProgram.CurrentDay
- [ ] Save user

### Completion Summary
- [ ] Show completion summary: "Workout logged successfully!"
- [ ] Show "Next workout: Day {N}"

### Helper Functions
- [ ] promptInt(prompt string) (int, error)
- [ ] buildCompletedWorkout(template *models.Workout, amrapReps map[models.LiftName]int) *models.Workout

### Testing
- [ ] Test successful logging flow
- [ ] Test AMRAP input validation
- [ ] Verify workout saved to history
- [ ] Verify CurrentDay increments
- [ ] Test with no current user/program

## Prompt 9: Advanced Workout Logging with Failure Mode

### Flag Addition
- [ ] Add --fail flag using cobra.Command.Flags().Bool("fail", false, "Record individual reps for each set")
- [ ] Access with cmd.Flags().GetBool("fail")

### Modified Collection Logic
- [ ] If --fail is false: Keep existing behavior (auto-complete except AMRAP)
- [ ] If --fail is true: Prompt for every set

### Failure Mode Prompting
- [ ] For each lift and set:
  ```
  Overhead Press - Set 1 (Warmup):
  Target: 5 reps @ 45 lbs
  How many reps completed? 
  ```
- [ ] Accept 0 for failed sets
- [ ] Validate non-negative integer
- [ ] Show set type (Warmup/Working/AMRAP)

### Helper Functions
- [ ] collectWithFailure(workout *models.Workout) *models.Workout
- [ ] Shows each set individually
- [ ] Collects actual reps for all sets

### Integration
- [ ] Keep existing save logic (same workout history update, same CurrentDay increment)

### Help Text Update
- [ ] Update help text:
  ```
  Log a completed workout. By default, assumes all non-AMRAP sets were completed successfully.
  Use --fail flag to record individual reps for each set.
  ```

### Testing
- [ ] Test default mode still works
- [ ] Test --fail flag changes behavior
- [ ] Test collecting reps for every set in fail mode
- [ ] Test 0 reps accepted for failures
- [ ] Verify both modes save correctly
- [ ] Test mixed success/failure sets

## Prompt 10: Progression System & Final Integration

### Progression Functions (workout/progression.go)
- [ ] Create AMRAP detection:
  - [ ] func GetAMRAPReps(lift *models.Lift) (int, error)
  - [ ] Find the AMRAP set (Type == models.AMRAPSet)
  - [ ] Return ActualReps
  - [ ] Error if no AMRAP set found

- [ ] Create weight calculation:
  - [ ] func CalculateNewWeight(currentWeight float64, amrapReps int, baseIncrement float64, rules *models.ProgressionRules) float64
  - [ ] If amrapReps < 5: return currentWeight * rules.DeloadPercentage
  - [ ] If amrapReps >= rules.DoubleThreshold: return currentWeight + (baseIncrement * 2)
  - [ ] Otherwise: return currentWeight + baseIncrement
  - [ ] Round down to 2.5 lbs

- [ ] Create full progression calculation:
  - [ ] func CalculateProgression(workout *models.Workout, currentWeights map[models.LiftName]float64, rules *models.ProgressionRules) (map[models.LiftName]float64, error)
  - [ ] For each lift in workout: get AMRAP reps, get base increment, calculate new weight
  - [ ] Return updated weights map

### Update Workout Logging
- [ ] After saving workout, before incrementing day:
  - [ ] Call progression.CalculateProgression()
  - [ ] Update UserProgram.CurrentWeights with new weights
  - [ ] Show weight changes to user:
    ```
    Weight Updates:
    Overhead Press: 95 → 97.5 lbs (+2.5)
    Squat: 135 → 140 lbs (+5)
    ```

### Weight Change Formatting
- [ ] Show old → new
- [ ] Show difference (+X or -X for deload)
- [ ] Use colors if available (green for increase, red for deload)

### Integration Updates
- [ ] Ensure CurrentWeights is used for next workout calculation
- [ ] Verify CurrentDay increments after progression

### Comprehensive Testing
- [ ] Test normal progression (AMRAP = 5-9)
- [ ] Test double progression (AMRAP >= 10)
- [ ] Test deload (AMRAP < 5)
- [ ] Test weight rounding in progression
- [ ] Test full integration: log workout → calculate progression → update weights → next workout uses new weights
- [ ] Test progression display formatting

### Final Integration Test
- [ ] Create user
- [ ] Start program with weights
- [ ] Log several workouts with different AMRAP counts
- [ ] Verify progression applied correctly
- [ ] Verify next workout uses updated weights

## General Testing Requirements

### Test Strategy
- [ ] Write tests FIRST (TDD approach)
- [ ] Use table-driven tests where appropriate
- [ ] Mock external dependencies
- [ ] Test both happy paths and error conditions
- [ ] Use temporary directories for file operations
- [ ] Clean up test artifacts

### Implementation Standards
- [ ] Progressive complexity - simple foundations first
- [ ] Early testing to catch issues before they compound
- [ ] Clear separation of concerns throughout
- [ ] Consistent error handling patterns
- [ ] User-friendly output at every step
- [ ] No orphaned code - everything gets integrated