package excel

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

// StudentImportInfo 学生导入信息（统一模板）
type StudentImportInfo struct {
	Username string // 学号/用户名（主键）
	Name     string // 姓名
	Class    string // 班级名称
	Password string // 初始密码（默认123456）
}

// ParseUnifiedImportExcel 解析统一导入模板（包含学生和班级）
func ParseUnifiedImportExcel(r io.Reader) ([]StudentImportInfo, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("无法读取Excel文件: %v", err)
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("Excel文件没有工作表")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("无法读取工作表: %v", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel文件没有数据（至少需要标题行和一行数据）")
	}

	var students []StudentImportInfo
	// 从第二行开始读取（跳过标题行）
	for i, row := range rows[1:] {
		if len(row) < 2 {
			continue
		}

		username := row[0]
		if username == "" {
			continue
		}

		name := ""
		if len(row) >= 2 {
			name = row[1]
		}

		class := ""
		if len(row) >= 3 {
			class = row[2]
		}

		password := "123456"
		if len(row) >= 4 && row[3] != "" {
			password = row[3]
		}

		students = append(students, StudentImportInfo{
			Username: username,
			Name:     name,
			Class:    class,
			Password: password,
		})

		// 限制最多导入1000名学生
		if i >= 1000 {
			break
		}
	}

	if len(students) == 0 {
		return nil, fmt.Errorf("Excel文件中没有有效的学生数据")
	}

	return students, nil
}

// CreateUnifiedTemplate 创建统一导入模板（学生+班级）
func CreateUnifiedTemplate() (*excelize.File, error) {
	f := excelize.NewFile()

	// 设置标题行
	f.SetCellValue("Sheet1", "A1", "学号/用户名（必填，主键）")
	f.SetCellValue("Sheet1", "B1", "姓名（必填）")
	f.SetCellValue("Sheet1", "C1", "班级名称（必填，不存在则自动创建）")
	f.SetCellValue("Sheet1", "D1", "初始密码（选填，默认123456）")

	// 设置示例数据
	f.SetCellValue("Sheet1", "A2", "2024001")
	f.SetCellValue("Sheet1", "B2", "张三")
	f.SetCellValue("Sheet1", "C2", "2024级软件工程1班")
	f.SetCellValue("Sheet1", "D2", "123456")

	f.SetCellValue("Sheet1", "A3", "2024002")
	f.SetCellValue("Sheet1", "B3", "李四")
	f.SetCellValue("Sheet1", "C3", "2024级软件工程1班")
	f.SetCellValue("Sheet1", "D3", "123456")

	f.SetCellValue("Sheet1", "A4", "2024003")
	f.SetCellValue("Sheet1", "B4", "王五")
	f.SetCellValue("Sheet1", "C4", "2024级软件工程2班")
	f.SetCellValue("Sheet1", "D4", "123456")

	// 设置列宽
	f.SetColWidth("Sheet1", "A", "A", 25)
	f.SetColWidth("Sheet1", "B", "B", 15)
	f.SetColWidth("Sheet1", "C", "C", 30)
	f.SetColWidth("Sheet1", "D", "D", 20)

	return f, nil
}

// ParseStudentsExcel 解析学生名单Excel（兼容旧格式）
func ParseStudentsExcel(r io.Reader) ([]StudentImportInfo, error) {
	return ParseUnifiedImportExcel(r)
}

// ParseClassesExcel 解析班级名单Excel（兼容旧格式，只返回班级名称列表）
func ParseClassesExcel(r io.Reader) ([]string, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("无法读取Excel文件: %v", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("Excel文件没有工作表")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("无法读取工作表: %v", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel文件没有数据")
	}

	var classes []string
	for i, row := range rows[1:] {
		if len(row) < 3 {
			continue
		}
		className := row[2] // 班级在第三列
		if className == "" {
			continue
		}
		// 去重
		exists := false
		for _, c := range classes {
			if c == className {
				exists = true
				break
			}
		}
		if !exists {
			classes = append(classes, className)
		}

		// 限制最多导入100个班级
		if i >= 100 {
			break
		}
	}

	return classes, nil
}

// CreateStudentTemplate 创建学生名单模板（兼容旧接口，使用统一模板）
func CreateStudentTemplate() (*excelize.File, error) {
	return CreateUnifiedTemplate()
}

// CreateClassTemplate 创建班级名单模板（兼容旧接口，使用统一模板）
func CreateClassTemplate() (*excelize.File, error) {
	return CreateUnifiedTemplate()
}
