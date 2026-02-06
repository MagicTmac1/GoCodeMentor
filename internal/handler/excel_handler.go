package handler

import (
	"GoCodeMentor/internal/pkg/excel"
	"GoCodeMentor/internal/service"
	"bytes"

	"github.com/gin-gonic/gin"
)

// ExcelHandler handles excel import/export requests.
type ExcelHandler struct {
	classSvc service.IClassService
	userSvc  service.IUserService
}

// NewExcelHandler creates a new ExcelHandler.
func NewExcelHandler(classSvc service.IClassService, userSvc service.IUserService) *ExcelHandler {
	return &ExcelHandler{classSvc: classSvc, userSvc: userSvc}
}

// DownloadStudentTemplate handles downloading the student list template.
func (h *ExcelHandler) DownloadStudentTemplate(c *gin.Context) {
	f, err := excel.CreateStudentTemplate()
	if err != nil {
		c.JSON(500, gin.H{"error": "创建模板失败"})
		return
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(500, gin.H{"error": "生成文件失败"})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=学生名单模板.xlsx")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// DownloadClassTemplate handles downloading the class list template.
func (h *ExcelHandler) DownloadClassTemplate(c *gin.Context) {
	f, err := excel.CreateClassTemplate()
	if err != nil {
		c.JSON(500, gin.H{"error": "创建模板失败"})
		return
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(500, gin.H{"error": "生成文件失败"})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=班级名单模板.xlsx")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// ImportStudentsToClass handles importing students into a class.
func (h *ExcelHandler) ImportStudentsToClass(c *gin.Context) {
	classID := c.Param("id")
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

	class, err := h.classSvc.GetClassByID(classID)
	if err != nil {
		c.JSON(404, gin.H{"error": "班级不存在"})
		return
	}

	if userRole != "teacher" || class.TeacherID != userID {
		c.JSON(403, gin.H{"error": "无权操作"})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "请选择Excel文件"})
		return
	}
	defer file.Close()

	students, err := excel.ParseStudentsExcel(file)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var successCount, failCount, createdCount int
	var failMessages []string

	for _, s := range students {
		var studentID string

		existingStudent, err := h.userSvc.GetByUsername(s.Username)
		if err != nil {
			name := s.Name
			if name == "" {
				name = s.Username
			}
			password := s.Password
			if password == "" {
				password = "123456"
			}

			newStudent, err := h.userSvc.Register(s.Username, password, name, "student")
			if err != nil {
				failCount++
				failMessages = append(failMessages, s.Username+": 创建用户失败 - "+err.Error())
				continue
			}
			studentID = newStudent.ID
			createdCount++
		} else {
			studentID = existingStudent.ID
		}

		if err := h.classSvc.AddStudentToClass(studentID, classID); err != nil {
			failCount++
			failMessages = append(failMessages, s.Username+": 添加到班级失败 - "+err.Error())
			continue
		}

		successCount++
	}

	c.JSON(200, gin.H{
		"success":       successCount,
		"failed":        failCount,
		"created":       createdCount,
		"total":         len(students),
		"fail_messages": failMessages,
	})
}

// ImportClassesAndStudents handles importing classes and students.
func (h *ExcelHandler) ImportClassesAndStudents(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

	if userRole != "teacher" {
		c.JSON(403, gin.H{"error": "只有教师可以导入班级"})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "请选择Excel文件"})
		return
	}
	defer file.Close()

	students, err := excel.ParseStudentsExcel(file)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	classMap := make(map[string]string) // className -> classID
	for _, s := range students {
		if s.Class != "" {
			classMap[s.Class] = ""
		}
	}

	var classSuccessCount, classFailCount int
	for className := range classMap {
		existingClasses, _ := h.classSvc.GetClassesByTeacherID(userID)
		exists := false
		for _, ec := range existingClasses {
			if ec.Name == className {
				classMap[className] = ec.ID
				exists = true
				break
			}
		}

		if !exists {
			class, err := h.classSvc.CreateClass(className, userID)
			if err != nil {
				classFailCount++
				continue
			}
			classMap[className] = class.ID
			classSuccessCount++
		}
	}

	var studentSuccessCount, studentFailCount, studentCreatedCount int
	var failMessages []string

	for _, s := range students {
		if s.Username == "" {
			continue
		}

		var studentID string

		existingStudent, err := h.userSvc.GetByUsername(s.Username)
		if err != nil {
			name := s.Name
			if name == "" {
				name = s.Username
			}
			password := s.Password
			if password == "" {
				password = "123456"
			}

			newStudent, err := h.userSvc.Register(s.Username, password, name, "student")
			if err != nil {
				studentFailCount++
				failMessages = append(failMessages, s.Username+": 创建用户失败 - "+err.Error())
				continue
			}
			studentID = newStudent.ID
			studentCreatedCount++
		} else {
			studentID = existingStudent.ID
		}

		if s.Class != "" {
			classID := classMap[s.Class]
			if classID != "" {
				if err := h.classSvc.AddStudentToClass(studentID, classID); err != nil {
					studentFailCount++
					failMessages = append(failMessages, s.Username+": 添加到班级失败 - "+err.Error())
					continue
				}
			}
		}

		studentSuccessCount++
	}

	c.JSON(200, gin.H{
		"class_success":   classSuccessCount,
		"class_failed":    classFailCount,
		"student_success": studentSuccessCount,
		"student_failed":  studentFailCount,
		"student_created": studentCreatedCount,
		"total":           len(students),
	})
}
