package catalog

import "planahead/planner-api/internal/model"

const WaterlooCode = "WATERLOO"

func ProgramDefinitions() []model.ProgramDefinition {
	return []model.ProgramDefinition{
		buildComputerScienceProgram(),
		buildMathematicsProgram(),
	}
}

func DemoProgress() map[string]model.ProgressSnapshot {
	return map[string]model.ProgressSnapshot{
		"CS": {
			CourseStatuses: map[string]model.CourseStatus{
				"CS 135":     model.CourseStatusCompleted,
				"MATH 135":   model.CourseStatusCompleted,
				"MATH 137":   model.CourseStatusCompleted,
				"COMMST 100": model.CourseStatusCompleted,
				"ECON 101":   model.CourseStatusCompleted,
				"CS 136":     model.CourseStatusInProgress,
				"MATH 136":   model.CourseStatusInProgress,
				"STAT 230":   model.CourseStatusInProgress,
				"PHYS 121":   model.CourseStatusPlanned,
				"CS 245":     model.CourseStatusPlanned,
				"CS 246":     model.CourseStatusPlanned,
			},
			ElectiveSelections: map[string]string{
				"Y1_FALL_BREADTH":   "ECON 101",
				"Y1_WINTER_SCIENCE": "PHYS 121",
			},
		},
		"MATH": {
			CourseStatuses: map[string]model.CourseStatus{
				"MATH 135": model.CourseStatusCompleted,
				"MATH 137": model.CourseStatusCompleted,
				"CS 115":   model.CourseStatusInProgress,
			},
			ElectiveSelections: map[string]string{
				"MATH_Y1_FALL_COMM": "ENGL 109",
			},
		},
	}
}

func university() model.University {
	return model.University{
		Code: WaterlooCode,
		Name: "University of Waterloo",
	}
}

func buildComputerScienceProgram() model.ProgramDefinition {
	return model.ProgramDefinition{
		UniversityCode: WaterlooCode,
		ProgramCode:    "CS",
		Name:           "Computer Science",
		Degree:         "Bachelor of Computer Science",
		Description:    "Waterloo Computer Science roadmap with core sequencing, electives, and prerequisite validation.",
		Terms: []model.TermDefinition{
			{
				Code:     "Y1_FALL",
				Label:    "Year 1 Fall",
				Year:     1,
				Season:   model.TermSeasonFall,
				Sequence: 1,
				Requirements: []model.TermRequirementDefinition{
					courseReq("CS 135", "Designing Functional Programs", "CS", "Introductory functional programming and algorithmic design.", 1, nil),
					courseReq("MATH 135", "Algebra for Honours Mathematics", "MATH", "Foundations of algebra and mathematical reasoning.", 2, nil),
					courseReq("MATH 137", "Calculus 1", "MATH", "Differential calculus for honours mathematics.", 3, nil),
					courseReq("COMMST 100", "Communication in Mathematics and Computer Science", "COMM", "Communication practices for technical fields.", 4, nil),
					groupReq("Y1_FALL_BREADTH", "Choose 1 breadth elective", "Pick one humanities or social science course to round out the first term.", 5,
						option("ECON 101", "Introduction to Microeconomics", "ECON", "Economic decision-making and market behaviour.", 1, nil),
						option("PSYCH 101", "Introductory Psychology", "PSYCH", "Survey of psychological science.", 2, nil),
						option("ENGL 109", "Introduction to Academic Writing", "ENGL", "Critical writing and argumentation.", 3, nil),
					),
				},
			},
			{
				Code:     "Y1_WINTER",
				Label:    "Year 1 Winter",
				Year:     1,
				Season:   model.TermSeasonWinter,
				Sequence: 2,
				Requirements: []model.TermRequirementDefinition{
					courseReq("CS 136", "Elementary Algorithm Design and Data Abstraction", "CS", "Continuation of CS 135 with data abstraction.", 1, []string{"CS 135"}),
					courseReq("MATH 136", "Linear Algebra 1", "MATH", "Linear systems, vector spaces, and matrices.", 2, []string{"MATH 135"}),
					courseReq("MATH 138", "Calculus 2", "MATH", "Integral calculus and series.", 3, []string{"MATH 137"}),
					courseReq("STAT 230", "Probability", "STAT", "Discrete and continuous probability for mathematical computing.", 4, []string{"MATH 137"}),
					groupReq("Y1_WINTER_SCIENCE", "Choose 1 science elective", "A lab-friendly science pick keeps the plan broad in first year.", 5,
						option("BIOL 130", "Cell Biology", "BIOL", "Cell structure, genetics, and biological systems.", 1, nil),
						option("CHEM 120", "General Chemistry 1", "CHEM", "Core chemistry concepts for science students.", 2, nil),
						option("PHYS 121", "Mechanics", "PHYS", "Mechanics and introductory physics problem solving.", 3, nil),
					),
				},
			},
			{
				Code:     "Y1_SPRING",
				Label:    "Year 1 Spring",
				Year:     1,
				Season:   model.TermSeasonSpring,
				Sequence: 3,
				Requirements: []model.TermRequirementDefinition{
					courseReq("CS 245", "Logic and Computation", "CS", "Discrete logic foundations for computer science.", 1, []string{"CS 136"}),
					courseReq("CS 246", "Object-Oriented Software Development", "CS", "Object-oriented design and software construction.", 2, []string{"CS 136"}),
					courseReq("MATH 239", "Introduction to Combinatorics", "MATH", "Counting, graphs, and combinatorial reasoning.", 3, []string{"MATH 136"}),
					groupReq("Y1_SPRING_COMM", "Choose 1 communication elective", "Pair core theory with a communication-focused elective.", 4,
						option("SPCOM 100", "Introduction to Speech Communication", "SPCOM", "Foundations of public speaking and communication.", 1, nil),
						option("ENGL 210F", "Genres of Creative Writing", "ENGL", "Writing workshop focused on creative genres.", 2, nil),
						option("LS 101", "Introduction to Legal Studies", "LS", "Survey of legal institutions and reasoning.", 3, nil),
					),
				},
			},
			{
				Code:     "Y2_FALL",
				Label:    "Year 2 Fall",
				Year:     2,
				Season:   model.TermSeasonFall,
				Sequence: 4,
				Requirements: []model.TermRequirementDefinition{
					courseReq("CS 240", "Data Structures and Data Management", "CS", "Core data structures and performance tradeoffs.", 1, []string{"CS 136"}),
					courseReq("CS 241", "Foundations of Sequential Programs", "CS", "Low-level program structure, compilation, and memory.", 2, []string{"CS 136"}),
					courseReq("STAT 231", "Statistics", "STAT", "Statistical inference and estimation.", 3, []string{"STAT 230"}),
					groupReq("Y2_FALL_CONTEXT", "Choose 1 context elective", "A business or history course adds context beyond core CS theory.", 4,
						option("AFM 101", "Introduction to Financial Accounting", "AFM", "Accounting and business reporting fundamentals.", 1, nil),
						option("BET 100", "Foundations of Venture Creation", "BET", "Innovation and venture ideation.", 2, nil),
						option("HIST 101", "World History Since 1500", "HIST", "Modern global historical developments.", 3, nil),
					),
				},
			},
			{
				Code:     "Y2_WINTER",
				Label:    "Year 2 Winter",
				Year:     2,
				Season:   model.TermSeasonWinter,
				Sequence: 5,
				Requirements: []model.TermRequirementDefinition{
					courseReq("CS 251", "Computer Organization and Design", "CS", "Computer architecture and systems fundamentals.", 1, []string{"CS 245"}),
					courseReq("CS 341", "Algorithms", "CS", "Algorithm design, analysis, and correctness.", 2, []string{"CS 240"}),
					courseReq("CS 350", "Operating Systems", "CS", "Concurrency, processes, and operating system internals.", 3, []string{"CS 246"}),
					groupReq("Y2_WINTER_HUMANITIES", "Choose 1 humanities elective", "A humanities elective keeps the roadmap broad and balanced.", 4,
						option("PHIL 145", "Critical Thinking", "PHIL", "Argument analysis and logical reasoning.", 1, nil),
						option("CLAS 104", "Classical Mythology", "CLAS", "Themes and narratives from classical antiquity.", 2, nil),
						option("HRM 200", "Basic Human Resources Management", "HRM", "Workplace management and HR foundations.", 3, nil),
					),
				},
			},
			{
				Code:     "Y2_SPRING",
				Label:    "Year 2 Spring",
				Year:     2,
				Season:   model.TermSeasonSpring,
				Sequence: 6,
				Requirements: []model.TermRequirementDefinition{
					courseReq("CS 348", "Introduction to Database Management", "CS", "Data modeling and relational database systems.", 1, []string{"CS 246"}),
					courseReq("CS 349", "User Interfaces", "CS", "User interface design and interactive systems.", 2, []string{"CS 246"}),
					courseReq("CS 370", "Numerical Computation", "CS", "Computational methods with numerical analysis.", 3, []string{"MATH 239"}),
					groupReq("Y2_SPRING_FREE", "Choose 1 open elective", "Keep one slot flexible for a specialization or broader interest.", 4,
						option("ECON 201", "Microeconomic Theory", "ECON", "Intermediate microeconomic analysis.", 1, nil),
						option("PSCI 150", "Politics and the State", "PSCI", "Political systems and governance.", 2, nil),
						option("MUSIC 140", "Understanding Music", "MUSIC", "Introductory music literacy and appreciation.", 3, nil),
					),
				},
			},
		},
	}
}

func buildMathematicsProgram() model.ProgramDefinition {
	return model.ProgramDefinition{
		UniversityCode: WaterlooCode,
		ProgramCode:    "MATH",
		Name:           "Mathematics",
		Degree:         "Bachelor of Mathematics",
		Description:    "Waterloo Mathematics roadmap with foundational algebra, calculus, and introductory computing.",
		Terms: []model.TermDefinition{
			{
				Code:     "MATH_Y1_FALL",
				Label:    "Year 1 Fall",
				Year:     1,
				Season:   model.TermSeasonFall,
				Sequence: 1,
				Requirements: []model.TermRequirementDefinition{
					courseReq("MATH 135", "Algebra for Honours Mathematics", "MATH", "Foundations of algebra and mathematical reasoning.", 1, nil),
					courseReq("MATH 137", "Calculus 1", "MATH", "Differential calculus for honours mathematics.", 2, nil),
					courseReq("CS 115", "Introduction to Computer Science 1", "CS", "Foundational programming for mathematical problem solving.", 3, nil),
					groupReq("MATH_Y1_FALL_COMM", "Choose 1 writing elective", "Strengthen written communication early in the degree.", 4,
						option("ENGL 109", "Introduction to Academic Writing", "ENGL", "Critical writing and argumentation.", 1, nil),
						option("COMMST 100", "Communication in Mathematics and Computer Science", "COMM", "Communication practices for technical fields.", 2, nil),
					),
				},
			},
			{
				Code:     "MATH_Y1_WINTER",
				Label:    "Year 1 Winter",
				Year:     1,
				Season:   model.TermSeasonWinter,
				Sequence: 2,
				Requirements: []model.TermRequirementDefinition{
					courseReq("MATH 136", "Linear Algebra 1", "MATH", "Linear systems, vector spaces, and matrices.", 1, []string{"MATH 135"}),
					courseReq("MATH 138", "Calculus 2", "MATH", "Integral calculus and series.", 2, []string{"MATH 137"}),
					courseReq("STAT 230", "Probability", "STAT", "Discrete and continuous probability for mathematical computing.", 3, []string{"MATH 137"}),
					courseReq("MATH 239", "Introduction to Combinatorics", "MATH", "Counting, graphs, and combinatorial reasoning.", 4, []string{"MATH 136"}),
				},
			},
		},
	}
}

func courseReq(code string, title string, subject string, description string, sequence int32, prerequisites []string) model.TermRequirementDefinition {
	course := model.CourseRequirementDefinition{
		Code:          toRequirementCode(code),
		Course:        course(code, title, subject, description),
		Sequence:      sequence,
		Prerequisites: toPrerequisites(prerequisites),
		Kind:          model.RequirementKindCourse,
	}
	return model.TermRequirementDefinition{
		Kind:     model.RequirementKindCourse,
		Sequence: sequence,
		Course:   &course,
	}
}

func groupReq(code string, title string, description string, sequence int32, options ...model.CourseRequirementDefinition) model.TermRequirementDefinition {
	groupDescription := description
	group := model.ElectiveGroupDefinition{
		Code:          code,
		Title:         title,
		Description:   description,
		Kind:          model.RequirementGroupKindOneOf,
		MinSelections: 1,
		MaxSelections: 1,
		Sequence:      sequence,
		Options:       options,
		Notes:         &groupDescription,
	}

	return model.TermRequirementDefinition{
		Kind:     model.RequirementKindElectiveGroup,
		Sequence: sequence,
		Group:    &group,
	}
}

func option(code string, title string, subject string, description string, sequence int32, prerequisites []string) model.CourseRequirementDefinition {
	return model.CourseRequirementDefinition{
		Code:          "OPTION_" + toRequirementCode(code),
		Course:        course(code, title, subject, description),
		Sequence:      sequence,
		Prerequisites: toPrerequisites(prerequisites),
		Kind:          model.RequirementKindElectiveGroup,
	}
}

func course(code string, title string, subject string, description string) model.CourseDefinition {
	subjectCopy := subject
	descriptionCopy := description
	return model.CourseDefinition{
		Code:        code,
		Title:       title,
		Credits:     0.5,
		Description: &descriptionCopy,
		Subject:     &subjectCopy,
	}
}

func toPrerequisites(codes []string) []model.PrerequisiteDefinition {
	if len(codes) == 0 {
		return nil
	}

	result := make([]model.PrerequisiteDefinition, 0, len(codes))
	for _, code := range codes {
		result = append(result, model.PrerequisiteDefinition{CourseCode: code})
	}
	return result
}

func toRequirementCode(courseCode string) string {
	result := make([]rune, 0, len(courseCode))
	for _, char := range courseCode {
		if char == ' ' {
			result = append(result, '_')
			continue
		}
		result = append(result, char)
	}
	return string(result)
}
