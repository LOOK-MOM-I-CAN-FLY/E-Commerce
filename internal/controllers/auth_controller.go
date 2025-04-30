package controllers

import (
	"digital-marketplace/internal/database"
	"digital-marketplace/internal/models"
	"digital-marketplace/internal/services"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Старые регулярные выражения (можно удалить, так как они теперь в ValidationService)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// Функции проверки (можно удалить, так как они теперь в ValidationService)
func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func isValidUsername(username string) bool {
	return usernameRegex.MatchString(username) && len(username) >= 3 && len(username) <= 30
}

// Moved from base_controller.go
// renderTemplate is a helper function to render HTML templates
// It automatically adds IsLoggedIn and User data to the template context if available
func renderTemplate(c *gin.Context, templateName string, data gin.H) {
	baseData := gin.H{
		"IsLoggedIn":      false, // Default to false
		"User":            nil,
		"Error":           nil, // Ensure Error is always available, default to nil
		"PasswordError":   nil, // Ensure PasswordError is always available
		"PasswordSuccess": nil, // Ensure PasswordSuccess is always available
	}

	// Check login status from context (set by AuthRequired or SetLoginStatus)
	if loggedInValue, exists := c.Get("is_logged_in"); exists {
		if isLoggedInBool, ok := loggedInValue.(bool); ok {
			baseData["IsLoggedIn"] = isLoggedInBool
		}
	}

	// Check user data from context
	if userValue, exists := c.Get("user"); exists {
		if userModel, ok := userValue.(models.User); ok {
			baseData["User"] = userModel
		}
	}

	// Merge provided data with base data
	// Provided data will overwrite base data if keys conflict (e.g., Error)
	for key, value := range data {
		baseData[key] = value
	}

	c.HTML(http.StatusOK, templateName, baseData)
}

// Moved from base_controller.go
// Helper function to get user from context
func getUserFromContext(c *gin.Context) (models.User, bool) {
	userValue, exists := c.Get("user")
	if !exists {
		return models.User{}, false
	}
	userModel, ok := userValue.(models.User)
	if !ok {
		// Handle error appropriately, maybe log it
		return models.User{}, false
	}
	return userModel, true
}

// Middleware to check if user is authenticated
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("user_id")
		if err != nil || cookie == "" {
			c.Set("is_logged_in", false)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		userID, err := strconv.ParseUint(cookie, 10, 64)
		if err != nil {
			// Invalid cookie value
			c.Set("is_logged_in", false)
			c.SetCookie("user_id", "", -1, "/", "", false, true) // Clear invalid cookie
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		var user models.User
		result := database.DB.First(&user, uint(userID))
		if result.Error != nil {
			// User not found in DB
			c.Set("is_logged_in", false)
			c.SetCookie("user_id", "", -1, "/", "", false, true) // Clear cookie for non-existent user
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Set user info and login status in context
		c.Set("is_logged_in", true)
		c.Set("user", user)

		c.Next()
	}
}

// Middleware to set login status for public pages
func SetLoginStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("user_id")
		isLoggedIn := false
		var currentUser models.User // Define user variable here
		userExists := false

		if err == nil && cookie != "" {
			userID, err := strconv.ParseUint(cookie, 10, 64)
			if err == nil {
				// We check if the user actually exists.
				if database.DB.First(&currentUser, uint(userID)).Error == nil {
					isLoggedIn = true
					userExists = true
				}
			}
		}
		c.Set("is_logged_in", isLoggedIn)
		if userExists {
			c.Set("user", currentUser) // Set user data in context if they exist
		}
		c.Next()
	}
}

type AuthController struct {
	oauthService      *services.OAuthService
	validationService *services.ValidationService
}

func NewAuthController() *AuthController {
	return &AuthController{
		oauthService:      services.NewOAuthService(),
		validationService: services.NewValidationService(),
	}
}

// ShowHome renders the index page
func (ac *AuthController) ShowHome(c *gin.Context) {
	renderTemplate(c, "index.html", gin.H{})
}

func (ac *AuthController) ShowRegister(c *gin.Context) {
	// Check if already logged in (using corrected type assertion)
	isLoggedIn := false
	if loggedInValue, exists := c.Get("is_logged_in"); exists {
		if loggedInBool, ok := loggedInValue.(bool); ok {
			isLoggedIn = loggedInBool
		}
	}

	if isLoggedIn {
		c.Redirect(http.StatusFound, "/profile")
		return
	}
	renderTemplate(c, "register.html", gin.H{})
}

func (ac *AuthController) Register(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	username := c.PostForm("username")

	// Валидация email с использованием ValidationService
	if valid, errMsg := ac.validationService.ValidateEmail(email); !valid {
		renderTemplate(c, "register.html", gin.H{
			"Error":    errMsg,
			"Username": username, // Возвращаем введенное имя пользователя для удобства
		})
		return
	}

	// Валидация имени пользователя с использованием ValidationService
	if valid, errMsg := ac.validationService.ValidateUsername(username); !valid {
		renderTemplate(c, "register.html", gin.H{
			"Error": errMsg,
			"Email": email, // Возвращаем введенный email для удобства
		})
		return
	}

	// Валидация пароля с использованием ValidationService
	if valid, errMsg := ac.validationService.ValidatePassword(password); !valid {
		renderTemplate(c, "register.html", gin.H{
			"Error":    errMsg,
			"Email":    email,
			"Username": username,
		})
		return
	}

	// Add validation (e.g., check if email exists)
	var existingUser models.User
	if database.DB.Where("email = ?", email).First(&existingUser).Error == nil {
		// User already exists
		renderTemplate(c, "register.html", gin.H{
			"Error":    "Пользователь с таким email уже существует",
			"Username": username,
		})
		return
	}

	// Проверка, не занято ли уже имя пользователя
	if database.DB.Where("username = ?", username).First(&existingUser).Error == nil {
		renderTemplate(c, "register.html", gin.H{
			"Error": "Это имя пользователя уже занято",
			"Email": email,
		})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := models.User{
		Email:     email,
		Username:  username, // Сохраняем имя пользователя
		Password:  string(hash),
		CreatedAt: time.Now(),
	}

	result := database.DB.Create(&user)
	if result.Error != nil {
		// Use renderTemplate to show error on the same page
		renderTemplate(c, "register.html", gin.H{
			"Error":    "Ошибка регистрации. Попробуйте снова.",
			"Email":    email,
			"Username": username,
		})
		return
	}

	c.Redirect(http.StatusFound, "/login") // Use StatusFound for redirects
}

func (ac *AuthController) ShowLogin(c *gin.Context) {
	// Check if already logged in (using corrected type assertion)
	isLoggedIn := false
	if loggedInValue, exists := c.Get("is_logged_in"); exists {
		if loggedInBool, ok := loggedInValue.(bool); ok {
			isLoggedIn = loggedInBool
		}
	}

	if isLoggedIn {
		c.Redirect(http.StatusFound, "/profile")
		return
	}
	renderTemplate(c, "login.html", gin.H{})
}

func (ac *AuthController) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	// Валидация email с использованием ValidationService
	if valid, errMsg := ac.validationService.ValidateEmail(email); !valid {
		renderTemplate(c, "login.html", gin.H{
			"Error": errMsg,
		})
		return
	}

	var user models.User
	result := database.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		log.Printf("%s: %v", c.Request.URL.Path, result.Error)
		// Use renderTemplate to show error
		renderTemplate(c, "login.html", gin.H{
			"Error": "Неверные учетные данные",
		})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Use renderTemplate to show error
		renderTemplate(c, "login.html", gin.H{
			"Error": "Неверные учетные данные",
		})
		return
	}

	// Use http.SameSiteLaxMode for broader compatibility
	c.SetCookie("user_id", fmt.Sprintf("%d", user.ID), 3600*24*7, "/", "", false, true) // Longer cookie duration (1 week)
	c.Redirect(http.StatusFound, "/profile")
}

func (ac *AuthController) Logout(c *gin.Context) {
	c.SetCookie("user_id", "", -1, "/", "", false, true) // Clear cookie
	c.Redirect(http.StatusFound, "/")
}

// ShowProfile renders the profile page
func (ac *AuthController) ShowProfile(c *gin.Context) {
	user, exists := getUserFromContext(c)
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Загружаем все товары, созданные пользователем
	var products []models.Product
	database.DB.Where("user_id = ?", user.ID).Find(&products)

	// Загружаем все заказы пользователя с присоединёнными товарами
	var orders []models.Order
	database.DB.Preload("Items").Preload("Items.Product").Where("user_id = ?", user.ID).Find(&orders)

	// Проверяем наличие сообщения об успешном заработке денег
	earnSuccess, _ := c.Get("earn_success")

	renderTemplate(c, "profile.html", gin.H{
		"Username":    user.Username,
		"Email":       user.Email,
		"Balance":     user.Balance,
		"Products":    products,
		"Orders":      orders,
		"EarnSuccess": earnSuccess,
	})
}

// ChangePassword handles the password change request
func (ac *AuthController) ChangePassword(c *gin.Context) {
	user, exists := getUserFromContext(c)
	if !exists {
		// This shouldn't happen if AuthRequired is used, but handle anyway
		c.Redirect(http.StatusFound, "/login")
		return
	}

	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmNewPassword := c.PostForm("confirm_new_password")

	// 1. Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		renderTemplate(c, "profile.html", gin.H{
			"PasswordError": "Текущий пароль неверен",
		})
		return
	}

	// 2. Check if new password and confirmation match
	if newPassword != confirmNewPassword {
		renderTemplate(c, "profile.html", gin.H{
			"PasswordError": "Новые пароли не совпадают",
		})
		return
	}

	// 3. Add validation for new password (length, complexity etc.) - SKIPPED for brevity

	// 4. Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		renderTemplate(c, "profile.html", gin.H{
			"PasswordError": "Ошибка при обработке нового пароля",
		})
		return
	}

	// 5. Update password in database
	result := database.DB.Model(&user).Update("password", string(newHash))
	if result.Error != nil {
		renderTemplate(c, "profile.html", gin.H{
			"PasswordError": "Не удалось обновить пароль в базе данных",
		})
		return
	}

	// 6. Redirect or show success message
	renderTemplate(c, "profile.html", gin.H{
		"PasswordSuccess": "Пароль успешно изменен",
	})
}

// InitiateGithubLogin redirects to GitHub for authentication
func (ac *AuthController) InitiateGithubLogin(c *gin.Context) {
	state, err := ac.oauthService.GenerateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}

	// Сохраняем state в cookie для проверки в callback
	c.SetCookie("oauth_state", state, 600, "/", "", false, true) // 10 минут

	// Перенаправляем на GitHub для авторизации
	url := ac.oauthService.GetGithubAuthURL(state)
	c.Redirect(http.StatusFound, url)
}

// HandleGithubCallback processes the GitHub OAuth callback
func (ac *AuthController) HandleGithubCallback(c *gin.Context) {
	// Получаем и проверяем state
	expectedState, err := c.Cookie("oauth_state")
	if err != nil || c.Query("state") != expectedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// Очищаем cookie state
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Получаем code из параметров запроса
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not provided"})
		return
	}

	// Обрабатываем код авторизации через сервис
	user, err := ac.oauthService.HandleGithubCallback(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("GitHub login failed: %v", err)})
		return
	}

	// Устанавливаем cookie с ID пользователя
	c.SetCookie("user_id", fmt.Sprintf("%d", user.ID), 3600*24*7, "/", "", false, true)
	c.Redirect(http.StatusFound, "/profile")
}

// EarnMoney позволяет пользователю заработать деньги
func (ac *AuthController) EarnMoney(c *gin.Context) {
	user, exists := getUserFromContext(c)
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Прибавляем 10 к балансу пользователя
	user.Balance += 10.0

	// Сохраняем обновленный баланс в базе данных
	if err := database.DB.Save(&user).Error; err != nil {
		// Если произошла ошибка, выводим ее в логи и перенаправляем на страницу профиля
		fmt.Printf("Ошибка при обновлении баланса пользователя %d: %v\n", user.ID, err)
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	// Устанавливаем сообщение об успешном заработке
	c.Set("earn_success", "Вы успешно заработали 10 кредитов!")

	// Перенаправляем обратно на страницу профиля
	c.Redirect(http.StatusFound, "/profile")
}
