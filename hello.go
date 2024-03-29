package main

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

func getLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "Не удалось получить IP-адрес"
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.To4() != nil && !ip.IsLoopback() {
				return ip.String()
			}
		}
	}
	return "IP-адрес не найден"
}

func getDiskIO() string {
	ioCounters, err := disk.IOCounters()
	if err != nil {
		return fmt.Sprintf("Ошибка при получении данных о вводе/выводе диска: %v", err)
	}

	var ioInfo []string
	for device, io := range ioCounters {
		info := fmt.Sprintf("Устройство: %s, Чтение: %v операций (%v GB), Запись: %v операций (%v GB)",
			device,
			io.ReadCount,
			float64(io.ReadBytes)/1024/1024/1024,
			io.WriteCount,
			float64(io.WriteBytes)/1024/1024/1024,
		)
		ioInfo = append(ioInfo, info)
	}

	if len(ioInfo) == 0 {
		return "Информация о вводе/выводе диска недоступна"
	}
	return strings.Join(ioInfo, "\n")
}

func main() {
	hostname, _ := os.Hostname()
	localIP := getLocalIP()
	currentUser, _ := user.Current()
	hInfo, _ := host.Info()
	vMem, _ := mem.VirtualMemory()
	cpus, _ := cpu.Info()
	cpuPercent, _ := cpu.Percent(time.Second, false)
	diskIO := getDiskIO()

	fileName := fmt.Sprintf("system_info_%s.txt", strings.ReplaceAll(hostname, " ", "_"))

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("Ошибка при создании файла: %s\n", err)
		return
	}
	defer file.Close()

	info := fmt.Sprintf(
		"Имя пользователя: %s\nИмя компьютера: %s\nIP-адрес: %s\nОбщее количество ОЗУ: %.2f GB\nИспользуемое ОЗУ: %.2f GB\nВерсия ОС: %s-%s\n",
		currentUser.Username,
		hostname,
		localIP,
		float64(vMem.Total)/1024/1024/1024,
		float64(vMem.Used)/1024/1024/1024,
		hInfo.Platform,
		hInfo.PlatformVersion,
	)
	if len(cpus) > 0 {
		info += fmt.Sprintf("CPU: %s, Ядер: %d\nИспользование CPU: %.2f%%\n", cpus[0].ModelName, cpus[0].Cores, cpuPercent[0])
	}
	info += fmt.Sprintf("Информация о дисковом вводе/выводе:\n%s\n", diskIO)

	_, err = file.WriteString(info)
	if err != nil {
		fmt.Printf("Ошибка при записи в файл: %s\n", err)
		return
	}

	fmt.Printf("Информация успешно записана в файл '%s'\n", fileName)
}
